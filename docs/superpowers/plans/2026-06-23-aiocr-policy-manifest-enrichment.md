# POLIS Manifest Insurance Enrichment — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Auto-fill the five insurance columns on each jamaah export row from a group umrah insurance policy (POLIS) PDF uploaded in the same scan session, joined per-person on the passport number.

**Architecture:** A router inside `ProcessDocumentsSync` splits uploaded files into identity (vision OCR → jamaah rows, unchanged) and policy (pdftotext → OpenCode text-mode LLM → `PolicyManifest`) lanes, then merges the manifest onto the identity rows by normalized passport number. Deterministic normalization (`mapAsuransi`, `normalizeDate`, `normPaspor`) and merge logic are unit-tested; the LLM extraction is validated live against the real sample PDF.

**Tech Stack:** Go, Fiber, OpenCode (OpenAI-compatible) chat API, poppler-utils `pdftotext` (already in the ai-ocr-service container), excelize.

## Global Constraints

- Module `github.com/jamaah-in/v2`. Package under test: `internal/aiocr/service`.
- Insurance values conform to Siskopatuh dropdowns: **Asuransi** → template Sheet2 list (18 values, embedded verbatim below); all dates → `yyyy-mm-dd`.
- Date semantics: Tgl Input Polis = certificate issue date; Tgl Awal Polis = keberangkatan (departure); Tgl Akhir Polis = kepulangan (return).
- Join key = passport number (manifest `No Identitas` ↔ row `no_paspor`), normalized = uppercase + all whitespace removed.
- Policy enrichment is OpenCode-only for v1 (`gemini` → extractor nil). Provider key empty → factory returns untyped nil (typed-nil guard, as in `NewAnalyzer`).
- New row-map keys: `asuransi`, `no_polis`, `tanggal_input_polis`, `tanggal_awal_polis`, `tanggal_akhir_polis`.
- No new Go dependency. No "Co-Authored-By"/AI trailer in commits.
- Run tests with `go test ./internal/aiocr/service/` (cgo off → no `-race` locally).

## File Structure

- **Create** `internal/aiocr/service/policy.go` — types (`PolicyEntry`, `PolicyManifest`, `PolicyExtractor`), `looksLikePolicy`, `extractPDFText` var+impl, `siskopatuhAsuransi` list, `mapAsuransi`, `normalizeDate`, `normPaspor`, `parsePolicyJSON`, `rowPasporKey`, `enrichRowsWithPolicy`, `policyPrompt`, `NewPolicyExtractor`.
- **Create** `internal/aiocr/service/policy_test.go` — unit tests for the deterministic pieces.
- **Modify** `internal/aiocr/service/opencode.go` — extract a shared `chat(ctx, content, maxTokens)` helper; add `(*OpenCodeAnalyzer) ExtractManifest`.
- **Modify** `internal/aiocr/service/export.go` — five insurance columns read from the row.
- **Modify** `internal/aiocr/service/export_test.go` — assert insurance cells when present.
- **Modify** `internal/aiocr/service/service.go` — `policy PolicyExtractor` field + `WithPolicy` setter.
- **Modify** `internal/aiocr/service/process_sync.go` — router + policy lane + merge.
- **Modify** `internal/aiocr/service/process_sync_test.go` — integration test with fake extractor.
- **Modify** `cmd/ai-ocr-service/main.go` — wire `NewPolicyExtractor`.

---

### Task 1: Policy value helpers (types, detection, normalizers)

**Files:**
- Create: `internal/aiocr/service/policy.go`
- Test: `internal/aiocr/service/policy_test.go`

**Interfaces:**
- Produces: `type PolicyEntry struct{ NoIdentitas, NoPolis, TanggalAwal, TanggalAkhir string }`; `type PolicyManifest struct{ Asuransi, TanggalInput string; Entries []PolicyEntry }`; `looksLikePolicy(string) bool`; `mapAsuransi(string) string`; `normalizeDate(string) string`; `normPaspor(string) string`. Uses `normUpper` from `siskopatuh_values.go`.

- [ ] **Step 1: Write the failing test**

```go
package service

import "testing"

func TestLooksLikePolicy(t *testing.T) {
	yes := []string{
		"CERTIFICATE TRAVEL INSURANCE\nPT. Asuransi Askrida Syariah",
		"MANIFEST JABAL RAHMAH ... NO POLIS",
		"Jenis Asuransi : ASURANSI PERJALANAN SYARIAH",
	}
	no := []string{"PROVINSI JAWA BARAT\nNIK 3273...\nPekerjaan", "REPUBLIK INDONESIA PASPOR"}
	for _, s := range yes {
		if !looksLikePolicy(s) {
			t.Errorf("looksLikePolicy(%q) = false, want true", s)
		}
	}
	for _, s := range no {
		if looksLikePolicy(s) {
			t.Errorf("looksLikePolicy(%q) = true, want false", s)
		}
	}
}

func TestMapAsuransi(t *testing.T) {
	cases := map[string]string{
		"PT. ASURANSI ASKRIDA SYARIAH": "ASURANSI ASKRIDA SYARIAH",
		"Asuransi Askrida Syariah":     "ASURANSI ASKRIDA SYARIAH",
		"ASKRIDA SYARIAH":              "ASURANSI ASKRIDA SYARIAH",
		"PT ASURANSI JASINDO SYARIAH":  "PT ASURANSI JASINDO SYARIAH",
		"SINARMAS SYARIAH":             "ASURANSI SINARMAS SYARIAH",
		"":                             "",
		"KOPERASI XYZ":                 "KOPERASI XYZ", // unknown -> cleaned passthrough
	}
	for in, want := range cases {
		if got := mapAsuransi(in); got != want {
			t.Errorf("mapAsuransi(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeDate(t *testing.T) {
	cases := map[string]string{
		"17-June-2026": "2026-06-17",
		"01-Jul-2026":  "2026-07-01",
		"2026-07-09":   "2026-07-09",
		"9/7/2026":     "2026-07-09",
		"":             "",
		"bukan tanggal": "bukan tanggal",
	}
	for in, want := range cases {
		if got := normalizeDate(in); got != want {
			t.Errorf("normalizeDate(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormPaspor(t *testing.T) {
	for _, in := range []string{"x2664222", "X2664222", " X2 664222 "} {
		if got := normPaspor(in); got != "X2664222" {
			t.Errorf("normPaspor(%q) = %q, want X2664222", in, got)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/aiocr/service/ -run 'TestLooksLikePolicy|TestMapAsuransi|TestNormalizeDate|TestNormPaspor'`
Expected: FAIL — undefined: `looksLikePolicy`, `mapAsuransi`, `normalizeDate`, `normPaspor`.

- [ ] **Step 3: Write minimal implementation**

Create `internal/aiocr/service/policy.go`:

```go
package service

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// PolicyEntry is one person's row from a POLIS manifest, keyed by passport.
type PolicyEntry struct {
	NoIdentitas  string // passport / ID number (join key)
	NoPolis      string
	TanggalAwal  string // yyyy-mm-dd (keberangkatan)
	TanggalAkhir string // yyyy-mm-dd (kepulangan)
}

// PolicyManifest is the extracted content of a POLIS PDF: document-level
// Asuransi + issue date, plus one entry per insured participant.
type PolicyManifest struct {
	Asuransi     string // insurer (document-level, raw)
	TanggalInput string // yyyy-mm-dd (document-level)
	Entries      []PolicyEntry
}

// PolicyExtractor turns a POLIS PDF's text into a structured manifest.
type PolicyExtractor interface {
	ExtractManifest(ctx context.Context, pdfText string) (*PolicyManifest, error)
}

// looksLikePolicy reports whether extracted PDF text is a travel-insurance
// policy/certificate (vs an identity document).
func looksLikePolicy(text string) bool {
	u := strings.ToUpper(text)
	for _, m := range []string{
		"CERTIFICATE TRAVEL INSURANCE", "ASURANSI PERJALANAN",
		"MANIFEST", "NOMOR POLIS", "NO POLIS",
	} {
		if strings.Contains(u, m) {
			return true
		}
	}
	return false
}

// extractPDFText is a package var so tests can stub it without invoking pdftotext.
var extractPDFText = extractPDFTextImpl

// extractPDFTextImpl renders a PDF to plain text via poppler's `pdftotext`
// (-layout keeps table columns roughly aligned; "-" writes to stdout).
func extractPDFTextImpl(ctx context.Context, data []byte) (string, error) {
	in, err := os.CreateTemp("", "polis-*.pdf")
	if err != nil {
		return "", err
	}
	defer os.Remove(in.Name())
	if _, err := in.Write(data); err != nil {
		in.Close()
		return "", err
	}
	in.Close()
	out, err := exec.CommandContext(ctx, "pdftotext", "-layout", in.Name(), "-").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// siskopatuhAsuransi is the template Sheet2 "Asuransi" dropdown (verbatim).
var siskopatuhAsuransi = []string{
	"AJS AMANAH JIWA GIRI ARTHA",
	"ASURANSI ASKRIDA SYARIAH",
	"ASURANSI BRINS",
	"ASURANSI CENTRAL ASIA SYARIAH",
	"ASURANSI CHUBB SYARIAH",
	"ASURANSI JIWA SYARIAH AL AMIN",
	"ASURANSI MAXIMUS GRAHA PERSADA UNIT SYARIAH",
	"ASURANSI RELIANCE INDONESIA UNIT SYARIAH",
	"ASURANSI SINARMAS SYARIAH",
	"ASURANSI SONWELIS TAKAFUL",
	"ASURANSI TAKAFUL UMUM",
	"ASURANSI TRI PAKARTA UNIT SYARIAH",
	"ASURANSI TUGU PRATAMA INDONESIA",
	"PAN PACIFIC SYARIAH INSURANCE",
	"PT ASURANSI JASINDO SYARIAH",
	"PT. ASURANSI UMUM MEGA UNIT SYARIAH",
	"SYARIAH BUMIDA",
	"ZURICH GENERAL TAKAFUL INDONESIA",
}

// mapAsuransi canonicalises an insurer name to the Sheet2 Asuransi dropdown,
// or returns the cleaned uppercase input when no confident match exists.
func mapAsuransi(raw string) string {
	t := normUpper(raw)
	t = strings.TrimSpace(strings.TrimPrefix(t, "PT."))
	t = strings.TrimSpace(strings.TrimPrefix(t, "PT"))
	t = normUpper(t)
	if t == "" {
		return ""
	}
	for _, a := range siskopatuhAsuransi {
		if normUpper(a) == t {
			return a
		}
	}
	for _, a := range siskopatuhAsuransi {
		ua := normUpper(a)
		// input contains a full canonical name, or a long input is contained
		// in a canonical name (e.g. "ASKRIDA SYARIAH" ⊂ "ASURANSI ASKRIDA SYARIAH").
		if strings.Contains(t, ua) || (len(t) >= 6 && strings.Contains(ua, t)) {
			return a
		}
	}
	return t
}

// normalizeDate parses the date formats a POLIS uses and returns yyyy-mm-dd;
// unparseable input is returned trimmed and unchanged.
func normalizeDate(s string) string {
	t := strings.TrimSpace(s)
	if t == "" {
		return ""
	}
	for _, l := range []string{
		"2006-01-02", "02-Jan-2006", "2-Jan-2006",
		"02-January-2006", "2-January-2006", "02/01/2006", "2/1/2006",
	} {
		if tm, err := time.Parse(l, t); err == nil {
			return tm.Format("2006-01-02")
		}
	}
	return t
}

// normPaspor is the passport join key: uppercase with all whitespace removed.
func normPaspor(s string) string {
	return strings.Join(strings.Fields(strings.ToUpper(s)), "")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/aiocr/service/ -run 'TestLooksLikePolicy|TestMapAsuransi|TestNormalizeDate|TestNormPaspor'`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/aiocr/service/policy.go internal/aiocr/service/policy_test.go
git commit -m "feat(aiocr): policy detection + Asuransi/date/paspor normalizers"
```

---

### Task 2: Manifest JSON parse + row merge

**Files:**
- Modify: `internal/aiocr/service/policy.go`
- Test: `internal/aiocr/service/policy_test.go`

**Interfaces:**
- Consumes: `PolicyManifest`, `PolicyEntry`, `normalizeDate`, `normPaspor`, `mapAsuransi` (Task 1).
- Produces: `parsePolicyJSON(string) (*PolicyManifest, error)`; `rowPasporKey(map[string]any) string`; `enrichRowsWithPolicy(data []any, entries map[string]PolicyEntry, asuransi, tglInput string)`.

- [ ] **Step 1: Write the failing test**

```go
func TestParsePolicyJSON(t *testing.T) {
	in := `{"asuransi":"PT. ASURANSI ASKRIDA SYARIAH","tanggal_input_polis":"17-June-2026",
	"peserta":[{"no_identitas":"X2664222","no_polis":"122015022600316-000043","tanggal_awal_polis":"01-Jul-2026","tanggal_akhir_polis":"09-Jul-2026"},
	{"no_identitas":" ","no_polis":"x"}]}`
	m, err := parsePolicyJSON(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.Asuransi != "PT. ASURANSI ASKRIDA SYARIAH" || m.TanggalInput != "2026-06-17" {
		t.Errorf("doc-level: %q / %q", m.Asuransi, m.TanggalInput)
	}
	if len(m.Entries) != 1 { // blank no_identitas dropped
		t.Fatalf("entries = %d, want 1", len(m.Entries))
	}
	e := m.Entries[0]
	if e.NoIdentitas != "X2664222" || e.NoPolis != "122015022600316-000043" ||
		e.TanggalAwal != "2026-07-01" || e.TanggalAkhir != "2026-07-09" {
		t.Errorf("entry = %+v", e)
	}
}

func TestEnrichRowsWithPolicy(t *testing.T) {
	rows := []any{
		map[string]any{"nama": "LESTARI", "no_paspor": "x2664222", "jenis_identitas": "PASPOR"},
		map[string]any{"nama": "BUDI", "no_identitas": "3273123456780001", "jenis_identitas": "NIK"},
	}
	entries := map[string]PolicyEntry{
		"X2664222": {NoIdentitas: "X2664222", NoPolis: "POL-43", TanggalAwal: "2026-07-01", TanggalAkhir: "2026-07-09"},
	}
	enrichRowsWithPolicy(rows, entries, "PT. ASURANSI ASKRIDA SYARIAH", "2026-06-17")

	r0 := rows[0].(map[string]any)
	if r0["asuransi"] != "ASURANSI ASKRIDA SYARIAH" || r0["no_polis"] != "POL-43" ||
		r0["tanggal_input_polis"] != "2026-06-17" || r0["tanggal_awal_polis"] != "2026-07-01" ||
		r0["tanggal_akhir_polis"] != "2026-07-09" {
		t.Errorf("matched row not enriched: %+v", r0)
	}
	r1 := rows[1].(map[string]any)
	if _, ok := r1["no_polis"]; ok {
		t.Errorf("unmatched row should be untouched: %+v", r1)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/aiocr/service/ -run 'TestParsePolicyJSON|TestEnrichRowsWithPolicy'`
Expected: FAIL — undefined: `parsePolicyJSON`, `enrichRowsWithPolicy`.

- [ ] **Step 3: Write minimal implementation**

Append to `internal/aiocr/service/policy.go` (add `"encoding/json"` to the import block):

```go
// parsePolicyJSON converts the LLM's manifest JSON into a PolicyManifest,
// normalising dates and dropping entries without an identity number.
func parsePolicyJSON(s string) (*PolicyManifest, error) {
	var raw struct {
		Asuransi     string `json:"asuransi"`
		TanggalInput string `json:"tanggal_input_polis"`
		Peserta      []struct {
			NoIdentitas  string `json:"no_identitas"`
			NoPolis      string `json:"no_polis"`
			TanggalAwal  string `json:"tanggal_awal_polis"`
			TanggalAkhir string `json:"tanggal_akhir_polis"`
		} `json:"peserta"`
	}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return nil, err
	}
	m := &PolicyManifest{
		Asuransi:     strings.TrimSpace(raw.Asuransi),
		TanggalInput: normalizeDate(raw.TanggalInput),
	}
	for _, p := range raw.Peserta {
		if strings.TrimSpace(p.NoIdentitas) == "" {
			continue
		}
		m.Entries = append(m.Entries, PolicyEntry{
			NoIdentitas:  strings.TrimSpace(p.NoIdentitas),
			NoPolis:      strings.TrimSpace(p.NoPolis),
			TanggalAwal:  normalizeDate(p.TanggalAwal),
			TanggalAkhir: normalizeDate(p.TanggalAkhir),
		})
	}
	return m, nil
}

// rowPasporKey is a jamaah row's passport join key: no_paspor, or no_identitas
// when the row's identity type is a passport.
func rowPasporKey(row map[string]any) string {
	get := func(k string) string {
		if v, ok := row[k]; ok && v != nil {
			return strings.TrimSpace(fmtAny(v))
		}
		return ""
	}
	if p := get("no_paspor"); p != "" {
		return normPaspor(p)
	}
	if strings.EqualFold(get("jenis_identitas"), "PASPOR") {
		return normPaspor(get("no_identitas"))
	}
	return ""
}

// enrichRowsWithPolicy fills the five insurance keys on every jamaah row whose
// passport matches a manifest entry. Document-level Asuransi is canonicalised to
// the Sheet2 dropdown. Rows with no match are left unchanged.
func enrichRowsWithPolicy(data []any, entries map[string]PolicyEntry, asuransi, tglInput string) {
	canonAsuransi := mapAsuransi(asuransi)
	for _, item := range data {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		key := rowPasporKey(row)
		if key == "" {
			continue
		}
		e, ok := entries[key]
		if !ok {
			continue
		}
		if canonAsuransi != "" {
			row["asuransi"] = canonAsuransi
		}
		if tglInput != "" {
			row["tanggal_input_polis"] = tglInput
		}
		if e.NoPolis != "" {
			row["no_polis"] = e.NoPolis
		}
		if e.TanggalAwal != "" {
			row["tanggal_awal_polis"] = e.TanggalAwal
		}
		if e.TanggalAkhir != "" {
			row["tanggal_akhir_polis"] = e.TanggalAkhir
		}
	}
}

// fmtAny stringifies a map value the same way the export's fieldGetter does.
func fmtAny(v any) string {
	switch s := v.(type) {
	case string:
		return s
	default:
		return strings.TrimSpace(fmtSprint(v))
	}
}
```

Add a tiny helper at the bottom of `policy.go` (avoids importing `fmt` twice with a clear name):

```go
import_fmt_note: add "fmt" to the import block, then:

func fmtSprint(v any) string { return fmt.Sprintf("%v", v) }
```

(Concretely: the `policy.go` import block becomes `"context"`, `"encoding/json"`, `"fmt"`, `"os"`, `"os/exec"`, `"strings"`, `"time"`.)

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/aiocr/service/ -run 'TestParsePolicyJSON|TestEnrichRowsWithPolicy'`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/aiocr/service/policy.go internal/aiocr/service/policy_test.go
git commit -m "feat(aiocr): manifest JSON parse + passport-keyed row enrichment"
```

---

### Task 3: Export the five insurance columns from the row

**Files:**
- Modify: `internal/aiocr/service/export.go:53-58` (the five blank insurance columns)
- Test: `internal/aiocr/service/export_test.go`

**Interfaces:**
- Consumes: `mapAsuransi` (Task 1), `fieldGetter.first` (existing).

- [ ] **Step 1: Write the failing test** — append to `export_test.go`:

```go
func TestExportFillsInsuranceColumns(t *testing.T) {
	records := []map[string]any{{
		"nama": "LESTARI EKA CITRA", "no_paspor": "X2664222", "jenis_identitas": "PASPOR",
		"asuransi": "ASURANSI ASKRIDA SYARIAH", "no_polis": "122015022600316-000043",
		"tanggal_input_polis": "2026-06-17", "tanggal_awal_polis": "2026-07-01",
		"tanggal_akhir_polis": "2026-07-09",
	}}
	data, err := generateInlineSiskopatuhExcel(records)
	if err != nil {
		t.Fatal(err)
	}
	f, _ := excelize.OpenReader(bytes.NewReader(data))
	cell := func(c string) string { v, _ := f.GetCellValue("Sheet1", c); return v }
	for c, want := range map[string]string{
		"AA2": "ASURANSI ASKRIDA SYARIAH", "AB2": "122015022600316-000043",
		"AC2": "2026-06-17", "AD2": "2026-07-01", "AE2": "2026-07-09", "AF2": "", // No BPJS blank
	} {
		if got := cell(c); got != want {
			t.Errorf("insurance %s = %q, want %q", c, got, want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/aiocr/service/ -run TestExportFillsInsuranceColumns`
Expected: FAIL — `AA2 = "" , want "ASURANSI ASKRIDA SYARIAH"` (columns still hard-coded blank).

- [ ] **Step 3: Write minimal implementation** — in `export.go`, replace these five entries:

```go
	{"Asuransi", func(g fieldGetter) string { return "" }},
	{"No Polis", func(g fieldGetter) string { return "" }},
	{"Tanggal Input Polis (yyyy-mm-dd)", func(g fieldGetter) string { return "" }},
	{"Tanggal Awal Polis (yyyy-mm-dd)", func(g fieldGetter) string { return "" }},
	{"Tanggal Akhir Polis (yyyy-mm-dd)", func(g fieldGetter) string { return "" }},
```

with:

```go
	{"Asuransi", func(g fieldGetter) string { return mapAsuransi(g.first("asuransi")) }},
	{"No Polis", func(g fieldGetter) string { return g.first("no_polis") }},
	{"Tanggal Input Polis (yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_input_polis") }},
	{"Tanggal Awal Polis (yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_awal_polis") }},
	{"Tanggal Akhir Polis (yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_akhir_polis") }},
```

Leave `{"No BPJS", func(g fieldGetter) string { return "" }}` unchanged.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/aiocr/service/ -run 'TestExport'`
Expected: PASS (both `TestExportFillsInsuranceColumns` and the existing `TestExportMatchesJamaahTemplate` — the latter still sees blank insurance because its records carry no insurance keys).

- [ ] **Step 5: Commit**

```bash
git add internal/aiocr/service/export.go internal/aiocr/service/export_test.go
git commit -m "feat(aiocr): export insurance columns from row data"
```

---

### Task 4: OpenCode text-mode `ExtractManifest` + factory

**Files:**
- Modify: `internal/aiocr/service/opencode.go`
- Modify: `internal/aiocr/service/policy.go` (add `policyPrompt`, `NewPolicyExtractor`)

**Interfaces:**
- Consumes: `OpenCodeAnalyzer`, `cleanJSONString` (existing); `parsePolicyJSON` (Task 2); `config.Config` (existing).
- Produces: `(*OpenCodeAnalyzer) ExtractManifest(ctx, pdfText) (*PolicyManifest, error)`; `NewPolicyExtractor(cfg *config.Config) PolicyExtractor`.

> No new unit test: `parsePolicyJSON` (Task 2) already covers the deterministic seam, and the live LLM call is validated in Task 6. This task is a refactor + thin wiring; its gate is `go build` + the existing suite staying green.

- [ ] **Step 1: Extract a shared `chat` helper in `opencode.go`**

Replace the body of `AnalyzeDocument` from the `reqBody := ...` block through the `raw, err := a.post(ctx, body)` + response parsing with a call to a new helper, and add the helper. Concretely, the request/response plumbing becomes:

```go
// chat sends one user message (content blocks already built) and returns the
// model's text reply (JSON fences stripped). Shared by vision OCR + policy text.
func (a *OpenCodeAnalyzer) chat(ctx context.Context, content []map[string]any, maxTokens int) (string, error) {
	reqBody := map[string]any{
		"model":       a.model,
		"temperature": 0.1,
		"max_tokens":  maxTokens,
		"messages":    []map[string]any{{"role": "user", "content": content}},
	}
	body, _ := json.Marshal(reqBody)
	raw, err := a.post(ctx, body)
	if err != nil {
		return "", err
	}
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("parse opencode response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("opencode api error: %s", resp.Error.Message)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("opencode returned no choices")
	}
	return cleanJSONString(resp.Choices[0].Message.Content), nil
}
```

And `AnalyzeDocument`'s tail becomes:

```go
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageData))
	content := []map[string]any{
		{"type": "text", "text": systemPrompts["auto"]},
		{"type": "image_url", "image_url": map[string]any{"url": dataURI}},
	}
	text, err := a.chat(ctx, content, 4096)
	if err != nil {
		return nil, err
	}
	var result OCRResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse extracted data: %w", err)
	}
	return &result, nil
```

- [ ] **Step 2: Add `ExtractManifest` + prompt + factory**

Append to `policy.go`:

```go
const policyPrompt = `Anda mengekstrak data dari SERTIFIKAT/POLIS asuransi perjalanan umrah Indonesia (grup).
Dari teks dokumen, kembalikan HANYA JSON dengan bentuk:
{
  "asuransi": "<nama perusahaan asuransi / Pengelola>",
  "tanggal_input_polis": "<tanggal terbit sertifikat, format yyyy-mm-dd>",
  "peserta": [
    {"no_identitas":"<nomor paspor/identitas>", "no_polis":"<nomor polis peserta>",
     "tanggal_awal_polis":"<tanggal keberangkatan yyyy-mm-dd>", "tanggal_akhir_polis":"<tanggal kepulangan yyyy-mm-dd>"}
  ]
}
Aturan: ambil daftar peserta dari tabel MANIFEST (kolom NO IDENTITAS, NO POLIS, TANGGAL KEBERANGKATAN, TANGGAL KEPULANGAN).
no_identitas = nomor paspor/identitas peserta. Semua tanggal format yyyy-mm-dd. Jika ragu suatu field, kosongkan. Tanpa teks lain selain JSON.`

// ExtractManifest implements PolicyExtractor using the OpenCode text chat API.
func (a *OpenCodeAnalyzer) ExtractManifest(ctx context.Context, pdfText string) (*PolicyManifest, error) {
	if !a.Available() {
		return nil, fmt.Errorf("opencode analyzer not configured (OPENCODE_API_KEY missing)")
	}
	if len(pdfText) > 60000 {
		pdfText = pdfText[:60000]
	}
	content := []map[string]any{
		{"type": "text", "text": policyPrompt + "\n\n=== TEKS DOKUMEN POLIS ===\n" + pdfText},
	}
	out, err := a.chat(ctx, content, 4096)
	if err != nil {
		return nil, err
	}
	return parsePolicyJSON(out)
}

// NewPolicyExtractor returns the configured PolicyExtractor, or nil when the
// provider is not opencode or its key is empty (callers treat nil as "no policy
// enrichment"). The key is checked before constructing, so a nil concrete
// pointer is never wrapped into a non-nil interface.
func NewPolicyExtractor(cfg *config.Config) PolicyExtractor {
	if cfg.AI.Provider == "opencode" && cfg.AI.OpenCodeAPIKey != "" {
		return NewOpenCodeAnalyzer(cfg.AI.OpenCodeAPIKey, cfg.AI.OpenCodeModel, cfg.AI.OpenCodeBaseURL)
	}
	return nil
}
```

Add `"github.com/jamaah-in/v2/internal/shared/config"` to `policy.go`'s import block.

- [ ] **Step 3: Build + run the whole package**

Run: `go build ./... && go test ./internal/aiocr/service/`
Expected: build clean; all existing + new tests PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/aiocr/service/opencode.go internal/aiocr/service/policy.go
git commit -m "feat(aiocr): OpenCode text-mode policy manifest extractor"
```

---

### Task 5: Router + policy lane + merge in `ProcessDocumentsSync`

**Files:**
- Modify: `internal/aiocr/service/service.go` (add field + setter)
- Modify: `internal/aiocr/service/process_sync.go`
- Test: `internal/aiocr/service/process_sync_test.go`

**Interfaces:**
- Consumes: `PolicyExtractor`, `extractPDFText`, `looksLikePolicy`, `enrichRowsWithPolicy`, `normPaspor` (Tasks 1-2).
- Produces: `(*AIOCRService).WithPolicy(PolicyExtractor) *AIOCRService`; `s.policy` field; revised `ProcessDocumentsSync`.

- [ ] **Step 1: Add the field + setter in `service.go`**

Change the struct and add a setter (no test yet — covered by Step 2's integration test):

```go
type AIOCRService struct {
	repo     *repository.AIOCRRepo
	analyzer DocumentAnalyzer
	policy   PolicyExtractor
	logger   *zap.SugaredLogger
}

// WithPolicy attaches a POLIS manifest extractor (enables insurance-column
// enrichment). Returns the receiver for chaining from main().
func (s *AIOCRService) WithPolicy(p PolicyExtractor) *AIOCRService {
	s.policy = p
	return s
}
```

- [ ] **Step 2: Write the failing integration test** — append to `process_sync_test.go`:

```go
// paspolAnalyzer returns a fixed passport number for every identity file.
type paspolAnalyzer struct{ paspor string }

func (p *paspolAnalyzer) Available() bool { return true }
func (p *paspolAnalyzer) AnalyzeDocument(ctx context.Context, data []byte, mime string) (*OCRResult, error) {
	return &OCRResult{DocType: "paspor", ExtractedData: ExtractedFields{NoPaspor: p.paspor}}, nil
}

type fakePolicy struct{ m *PolicyManifest }

func (f *fakePolicy) ExtractManifest(ctx context.Context, text string) (*PolicyManifest, error) {
	return f.m, nil
}

func TestProcessDocumentsSyncEnrichesFromPolicy(t *testing.T) {
	// Stub pdftotext: the policy file's bytes -> policy text; anything else -> "".
	orig := extractPDFText
	extractPDFText = func(ctx context.Context, data []byte) (string, error) {
		if string(data) == "POLISBYTES" {
			return "MANIFEST ... NO POLIS", nil
		}
		return "", nil
	}
	defer func() { extractPDFText = orig }()

	manifest := &PolicyManifest{
		Asuransi: "PT. ASURANSI ASKRIDA SYARIAH", TanggalInput: "2026-06-17",
		Entries: []PolicyEntry{{NoIdentitas: "X2664222", NoPolis: "POL-43",
			TanggalAwal: "2026-07-01", TanggalAkhir: "2026-07-09"}},
	}
	svc := (&AIOCRService{analyzer: &paspolAnalyzer{paspor: "X2664222"}, logger: zap.NewNop().Sugar()}).
		WithPolicy(&fakePolicy{m: manifest})

	files := []SyncFile{
		{FileName: "paspor.jpg", ContentType: "image/jpeg", Data: []byte("img")},
		{FileName: "polis.pdf", ContentType: "application/pdf", Data: []byte("POLISBYTES")},
	}
	res, err := svc.ProcessDocumentsSync(context.Background(), files, "default")
	if err != nil {
		t.Fatalf("ProcessDocumentsSync: %v", err)
	}
	if len(res.Data) != 1 {
		t.Fatalf("expected 1 jamaah row (policy makes no row), got %d", len(res.Data))
	}
	row := res.Data[0].(map[string]any)
	if row["no_polis"] != "POL-43" || row["asuransi"] != "ASURANSI ASKRIDA SYARIAH" ||
		row["tanggal_input_polis"] != "2026-06-17" || row["tanggal_awal_polis"] != "2026-07-01" {
		t.Errorf("row not enriched from policy: %+v", row)
	}
	// The policy file is reported, not turned into a jamaah row.
	var sawPolis bool
	for _, fr := range res.FileResults {
		if fr.Filename == "polis.pdf" {
			sawPolis = true
			if fr.Status != "completed" || fr.DocType != "polis" {
				t.Errorf("polis file_result = %+v", fr)
			}
		}
	}
	if !sawPolis {
		t.Errorf("polis.pdf missing from file_results")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/aiocr/service/ -run TestProcessDocumentsSyncEnrichesFromPolicy`
Expected: FAIL — the policy PDF is currently sent to the identity analyzer (becomes a row / wrong file_result), `res.Data` has 2 entries, `WithPolicy`/routing not present.

- [ ] **Step 4: Implement the router + merge in `process_sync.go`**

Add a classification pass before the worker loop, route policy files out of the identity batch, process them after the identity `wg.Wait()`, and enrich. Replace the function body from `if s.analyzer == nil { ... }` onward with:

```go
	if s.analyzer == nil {
		return nil, ErrOCRUnavailable
	}

	res := &ProcessDocumentsResult{
		Data:               []any{},
		ValidationWarnings: []any{},
		FileResults:        []SyncFileResult{},
	}

	// Classify each file: PDFs whose text looks like a POLIS go to the policy
	// lane; everything else is an identity document. pdftotext runs at most once
	// per PDF here and the text is reused for extraction.
	type policyDoc struct {
		file SyncFile
		text string
	}
	var identity []SyncFile
	var policies []policyDoc
	for _, f := range files {
		mime := f.ContentType
		if mime == "" {
			mime = detectMimeType(f.FileName)
		}
		if mime == "application/pdf" {
			if text, err := extractPDFText(ctx, f.Data); err == nil && looksLikePolicy(text) {
				policies = append(policies, policyDoc{file: f, text: text})
				continue
			}
		}
		identity = append(identity, f)
	}

	// --- identity lane (unchanged concurrency model) ---
	type fileOut struct {
		data       any
		hasData    bool
		warnings   []map[string]any
		fileResult SyncFileResult
	}
	outs := make([]fileOut, len(identity))

	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	for i := range identity {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			f := identity[i]
			mimeType := f.ContentType
			if mimeType == "" {
				mimeType = detectMimeType(f.FileName)
			}
			result, err := s.analyzer.AnalyzeDocument(ctx, f.Data, mimeType)
			if err != nil {
				s.logger.Errorf("ocr analyze %s: %v", f.FileName, err)
				outs[i].fileResult = SyncFileResult{
					Filename: f.FileName,
					Status:   "failed",
					Error:    "gagal mengekstrak dokumen (coba foto yang lebih jelas/terang)",
				}
				return
			}
			outs[i].data = normalizeToSiskopatuh(result.ExtractedData, result.DocType)
			outs[i].hasData = true
			for _, ve := range validateExtractedData(result.ExtractedData, result.DocType) {
				outs[i].warnings = append(outs[i].warnings, map[string]any{
					"filename": f.FileName,
					"field":    ve.Field,
					"message":  ve.Message,
					"value":    ve.Value,
				})
			}
			outs[i].fileResult = SyncFileResult{
				Filename: f.FileName,
				Status:   "completed",
				DocType:  result.DocType,
			}
		}(i)
	}
	wg.Wait()

	for i := range outs {
		if outs[i].hasData {
			res.Data = append(res.Data, outs[i].data)
		}
		for _, w := range outs[i].warnings {
			res.ValidationWarnings = append(res.ValidationWarnings, w)
		}
		res.FileResults = append(res.FileResults, outs[i].fileResult)
	}

	// --- policy lane: extract manifests, then enrich identity rows by passport ---
	entriesByPaspor := map[string]PolicyEntry{}
	var docAsuransi, docTglInput string
	for _, pd := range policies {
		if s.policy == nil {
			res.FileResults = append(res.FileResults, SyncFileResult{
				Filename: pd.file.FileName, Status: "failed",
				Error: "ekstraksi polis tidak tersedia (AI provider belum dikonfigurasi)",
			})
			continue
		}
		m, err := s.policy.ExtractManifest(ctx, pd.text)
		if err != nil || m == nil {
			s.logger.Errorf("policy extract %s: %v", pd.file.FileName, err)
			res.FileResults = append(res.FileResults, SyncFileResult{
				Filename: pd.file.FileName, Status: "failed",
				Error: "gagal membaca data polis (pastikan PDF asli, bukan hasil scan foto)",
			})
			continue
		}
		if m.Asuransi != "" {
			docAsuransi = m.Asuransi
		}
		if m.TanggalInput != "" {
			docTglInput = m.TanggalInput
		}
		for _, e := range m.Entries {
			entriesByPaspor[normPaspor(e.NoIdentitas)] = e
		}
		res.FileResults = append(res.FileResults, SyncFileResult{
			Filename: pd.file.FileName, Status: "completed", DocType: "polis",
		})
	}
	if len(entriesByPaspor) > 0 {
		enrichRowsWithPolicy(res.Data, entriesByPaspor, docAsuransi, docTglInput)
	}

	return res, nil
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/aiocr/service/`
Expected: PASS — new enrichment test + existing `TestProcessDocumentsSyncConcurrentAndOrdered` (still 10 identity files, no policy) stay green.

- [ ] **Step 6: Commit**

```bash
git add internal/aiocr/service/service.go internal/aiocr/service/process_sync.go internal/aiocr/service/process_sync_test.go
git commit -m "feat(aiocr): route POLIS PDFs to manifest extractor + enrich rows"
```

---

### Task 6: Wire main.go, full verify, live-validate, deploy

**Files:**
- Modify: `cmd/ai-ocr-service/main.go:64`

**Interfaces:**
- Consumes: `service.NewPolicyExtractor`, `(*AIOCRService).WithPolicy` (Tasks 4-5).

- [ ] **Step 1: Wire the extractor in `main.go`**

Replace:

```go
	aiocrService := service.NewAIOCRService(aiocrRepo, analyzer, logger)
```

with:

```go
	aiocrService := service.NewAIOCRService(aiocrRepo, analyzer, logger).
		WithPolicy(service.NewPolicyExtractor(cfg))
```

- [ ] **Step 2: Full build + vet + test**

Run: `gofmt -w internal/aiocr/service/*.go cmd/ai-ocr-service/main.go && go build ./... && go vet ./internal/aiocr/... && go test ./internal/aiocr/...`
Expected: all clean / PASS.

- [ ] **Step 3: Commit**

```bash
git add cmd/ai-ocr-service/main.go
git commit -m "feat(aiocr): enable POLIS insurance enrichment in ai-ocr-service"
```

- [ ] **Step 4: Live-validate the extractor against the real sample PDF**

Extract the sample PDF text and POST it to the OpenCode chat endpoint using the server's key (mirrors how the vision OCR was validated). From a shell with the key in `$OC_KEY`:

```bash
F="C:\\Users\\User\\Downloads\\POLIS - YELLOW - JABAL RAHMAH ANUGERAH 01 - 09 JUNI 2026 22PAX (1).pdf"
pdftotext -layout "$F" - | head -c 60000 > /tmp/polis.txt
# Build a chat request (model claude-haiku-4-5) embedding policyPrompt + /tmp/polis.txt,
# POST to https://opencode.ai/zen/v1/chat/completions, and confirm the JSON has
# asuransi="...ASKRIDA SYARIAH", tanggal_input_polis="2026-06-17", and ~22 peserta
# with the correct passports/policy numbers/dates (X2664222 -> ...-000043, 2026-07-01/09).
```

Expected: JSON parses; `peserta` contains the manifest rows with correct passports, policy numbers, and dates. If the model mis-reads, adjust `policyPrompt` and re-run before deploying.

- [ ] **Step 5: PR, merge, deploy**

```bash
git push -u origin feat/aiocr-policy-manifest-enrichment
gh pr create --title "POLIS manifest insurance enrichment (auto-fill Asuransi/No Polis/dates)" --body "<summary>"
gh pr merge --merge --delete-branch
```

Then on jni-server:

```bash
ssh jni-server-lan 'cd /data/docker/suluk && git pull --ff-only origin main \
  && docker compose -f deployments/docker-compose.yml build ai-ocr-service \
  && docker compose -f deployments/docker-compose.yml up -d ai-ocr-service'
```

Verify: `docker compose ps ai-ocr-service` shows Up; logs show `listening on :50056`. Ask the user to scan identity docs + the POLIS PDF together and export.

---

## Self-Review

**1. Spec coverage:**
- Detection/router → Task 5 (classify pass) + `looksLikePolicy` Task 1. ✓
- pdftotext extraction → `extractPDFText` Task 1. ✓
- `PolicyExtractor` (OpenCode text) + factory → Task 4. ✓
- Normalization (`mapAsuransi`/`normalizeDate`/`normPaspor`) → Task 1. ✓
- Manifest parse + merge → Task 2. ✓
- Export wiring (5 columns) → Task 3. ✓
- `AIOCRService.policy` + main wiring → Tasks 5-6. ✓
- Edge cases (no match untouched; `s.policy==nil`; malformed JSON → failed file_result) → Tasks 2 & 5. ✓
- Testing strategy (deterministic units + integration + live validation) → Tasks 1-6. ✓
- Out-of-scope items unimplemented by design. ✓

**2. Placeholder scan:** No TBD/TODO; every code step shows complete code. The live-validation step (Task 6.4) is prose because it is a one-off manual API check, not shipped code. ✓

**3. Type consistency:** `PolicyManifest{Asuransi, TanggalInput, Entries}` and `PolicyEntry{NoIdentitas, NoPolis, TanggalAwal, TanggalAkhir}` used identically across Tasks 1,2,4,5. Row keys `asuransi/no_polis/tanggal_input_polis/tanggal_awal_polis/tanggal_akhir_polis` identical in Tasks 2,3,5. `chat(ctx, content, maxTokens)` defined Task 4, used by `AnalyzeDocument` + `ExtractManifest`. `WithPolicy`/`NewPolicyExtractor` consistent Tasks 4-6. ✓

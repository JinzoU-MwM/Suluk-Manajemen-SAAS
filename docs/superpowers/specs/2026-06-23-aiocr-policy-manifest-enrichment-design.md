# AI-OCR Insurance Policy (POLIS) Manifest Enrichment — Design

**Date:** 2026-06-23
**Status:** Approved (design)
**Area:** `internal/aiocr` (ai-ocr-service)

## Goal

When an operator uploads an Indonesian group umrah **travel-insurance policy PDF**
(POLIS) together with the jamaah identity documents in a single scan session,
automatically fill the template's insurance columns on each matching jamaah row
by joining the policy's per-person **manifest** on the **passport number**.

Today those columns (`Asuransi`, `No Polis`, `Tanggal Input Polis`,
`Tanggal Awal Polis`, `Tanggal Akhir Polis`) are exported blank. After this
change they are populated from the POLIS for every jamaah present in the policy
manifest.

## Background — anatomy of a POLIS PDF

Sample: `POLIS - YELLOW - JABAL RAHMAH ANUGERAH ... 22PAX.pdf` (PT. Asuransi
Askrida Syariah, 40 pages, digitally generated text PDF).

- **Page 1 (cover):** `Pengelola/Operator: PT. ASURANSI ASKRIDA SYARIAH` (the
  insurer). `Nama Peserta` and `Nomor Polis` both read *"Terlampir"* (attached).
- **Page 3:** issue date `Jakarta, 17-June-2026` → **Tanggal Input Polis**.
- **Page 4 (MANIFEST):** one row per person:
  `NO | NAMA | NO IDENTITAS | TANGGAL LAHIR | USIA | NO POLIS | TGL KEBERANGKATAN | TGL KEPULANGAN`.
  The **No Identitas is the passport number** (e.g. `X2664222`) — identical to
  the value the passport scan produces (`no_paspor`). This is the join key.

Two data scopes:
- **Document-level** (same for every row): `Asuransi`, `Tanggal Input Polis`.
- **Per-person** (manifest, keyed by passport): `No Polis`, `Tanggal Awal Polis`
  (= keberangkatan/departure), `Tanggal Akhir Polis` (= kepulangan/return).

These mappings match the government template's own sample data exactly
(`Tanggal Input = 2026-06-17`, `Awal = 2026-07-01`, `Akhir = 2026-07-09`).

## Architecture

One scan request (`POST /api/process-documents`, multipart) already carries all
files. We add a **router** inside `ProcessDocumentsSync` that splits files into
two lanes, then **merges** the policy data onto the identity rows before
returning. No database, no new endpoint, no persistence — everything happens in
the single request, and the enriched preview rows the frontend holds are what
later get posted back to `/generate-excel`.

```
files ─┬─ identity (KTP/paspor)  ─► vision OCR ─► normalize ─► jamaah rows ─┐
       │                                                                     ├─ merge by passport ─► enriched rows
       └─ policy (POLIS PDF)      ─► pdftotext ─► LLM text ─► PolicyManifest ─┘
```

### Components

**1. Detection / router** — `internal/aiocr/service/policy.go`
- For each PDF file, run `pdftotext` once. If the extracted text contains POLIS
  markers, it is a policy file; otherwise it is an identity file. Non-PDF files
  (image/jpeg, image/png, …) are always identity files.
- `looksLikePolicy(text string) bool` — true when the text contains any of (case
  -insensitive): `"CERTIFICATE TRAVEL INSURANCE"`, `"ASURANSI PERJALANAN"`,
  `"MANIFEST"`, `"NOMOR POLIS"`, `"NO POLIS"`.
- `extractPDFText` is a package var (`= extractPDFTextImpl`) wrapping
  `pdftotext -layout <in> -` so tests can stub it. Mirrors the existing
  `rasterizePDF` pattern.

**2. Policy extraction** — `PolicyExtractor` interface
```go
type PolicyEntry struct {
    NoIdentitas  string // passport / ID number (join key)
    NoPolis      string
    TanggalAwal  string // yyyy-mm-dd (keberangkatan)
    TanggalAkhir string // yyyy-mm-dd (kepulangan)
}
type PolicyManifest struct {
    Asuransi      string        // insurer (document-level)
    TanggalInput  string        // yyyy-mm-dd (document-level)
    Entries       []PolicyEntry
}
type PolicyExtractor interface {
    ExtractManifest(ctx context.Context, pdfText string) (*PolicyManifest, error)
}
```
- Implemented on the OpenCode client (text-mode chat completion — **no vision**,
  cheaper and more accurate for a table). The prompt instructs the model to
  return strict JSON: `{asuransi, tanggal_input_polis, peserta:[{no_identitas,
  no_polis, tanggal_awal_polis, tanggal_akhir_polis}]}`, dates as `yyyy-mm-dd`.
- We reuse the OpenCode HTTP plumbing by extracting a shared
  `chat(ctx, content)` helper used by both the vision `AnalyzeDocument` and the
  text `ExtractManifest`.
- Factory `NewPolicyExtractor(cfg) PolicyExtractor`, mirroring `NewAnalyzer`.
  Returns nil when the provider key is empty or provider is `gemini` (policy
  enrichment is opencode-only for v1; the typed-nil guard from `NewAnalyzer`
  applies). `AIOCRService` gains a `policy PolicyExtractor` field set in
  `cmd/ai-ocr-service/main.go`.
- To bound tokens, only the first 60 000 chars of the PDF text are sent (the
  cover, notes, and manifest are early; trailing per-person certificate pages
  are boilerplate and safely trimmed).

**3. Normalization** (deterministic, fully tested)
- `mapAsuransi(raw string) string` — canonicalize to the Sheet2 `Asuransi`
  dropdown (18 values). Strip a leading `PT.`/`PT`, uppercase, collapse spaces;
  match against the list by normalized-equality, then by "list value is a
  substring of input or vice-versa" (so `PT. ASURANSI ASKRIDA SYARIAH` →
  `ASURANSI ASKRIDA SYARIAH`). No confident match → cleaned uppercase input
  (best-effort, never silently dropped). The 18 canonical values are embedded.
- `normalizeDate(s string) string` — parse `yyyy-mm-dd`, `dd-Mon-yyyy`
  (`17-June-2026`, `01-Jul-2026`), and `dd/mm/yyyy` → `yyyy-mm-dd`; unparseable
  → cleaned input. Applied defensively to all three policy dates.
- `normPaspor(s string) string` — uppercase + remove all whitespace; the join
  key on both sides.

**4. Merge** — `enrichRowsWithPolicy(rows []map[string]any, m *PolicyManifest)`
- Build `entriesByPaspor := map[normPaspor]PolicyEntry`.
- For each jamaah row, compute its passport key from `no_paspor` (fallback
  `no_identitas` when `jenis_identitas == "PASPOR"`). If an entry matches, set on
  the row: `asuransi` (= `mapAsuransi(m.Asuransi)`), `no_polis`,
  `tanggal_input_polis` (= `m.TanggalInput`), `tanggal_awal_polis`,
  `tanggal_akhir_polis`. Rows with no manifest match are left unchanged.
- Multiple policy files in one session merge into one `entriesByPaspor` (later
  entries win on key collision).

**5. Export wiring** — `internal/aiocr/service/export.go`
- Replace the five hard-coded `""` insurance columns with reads from the row:
  `Asuransi → mapAsuransi(g.first("asuransi"))`, `No Polis → g.first("no_polis")`,
  `Tanggal Input Polis → g.first("tanggal_input_polis")`,
  `Tanggal Awal Polis → g.first("tanggal_awal_polis")`,
  `Tanggal Akhir Polis → g.first("tanggal_akhir_polis")`. `No BPJS` stays blank.

### `ProcessDocumentsSync` flow (revised)

1. Partition `files` into `identityFiles` and `policyFiles` using the router
   (PDF text sniff; images always identity). The router keeps each policy file's
   already-extracted text to avoid a second `pdftotext`.
2. Process `identityFiles` concurrently exactly as today → `res.Data` rows +
   `file_results` + warnings.
3. For each policy file: `ExtractManifest(text)` → accumulate entries +
   document-level fields. Report the policy file in `file_results` with
   `doc_type:"polis"` and status `completed` (or `failed` with a clear message
   if extraction errors). A policy file produces **no** jamaah row.
4. If any manifest entries were collected, `enrichRowsWithPolicy(res.Data, …)`.
5. Return as today (now-enriched rows).

If `s.policy == nil` (no extractor configured) policy files are reported failed
with a clear message and identity processing proceeds unaffected.

## Error handling & edge cases

- **Jamaah not in manifest / manifest person not scanned:** no match → insurance
  blank for that row; no error. (Expected: children or late additions.)
- **Policy is a scanned image (no extractable text):** `looksLikePolicy` is
  false → the file is treated as an identity doc and will simply fail identity
  OCR with the normal "foto kurang jelas" message. Vision-based policy reading is
  **out of scope** (see below).
- **Unknown insurer name:** `mapAsuransi` passes the cleaned name through so the
  operator can fix it; it does not block the row.
- **Passport-less jamaah (KTP only):** no passport key → not joined. Acceptable
  (umrah manifests are keyed by passport).
- **LLM returns malformed JSON:** extraction returns an error → that policy file
  is `failed`; identity rows still returned un-enriched.

## Testing strategy

Deterministic units (table tests):
- `looksLikePolicy` — positive (cover/manifest text) and negative (KTP text).
- `mapAsuransi` — `PT. ASURANSI ASKRIDA SYARIAH` → `ASURANSI ASKRIDA SYARIAH`;
  `Askrida Syariah` → `ASURANSI ASKRIDA SYARIAH`; unknown → cleaned passthrough;
  every output is in the Sheet2 set or equals the cleaned input.
- `normalizeDate` — `17-June-2026`→`2026-06-17`, `01-Jul-2026`→`2026-07-01`,
  `2026-07-09`→`2026-07-09`, junk→cleaned input.
- `normPaspor` — case/space normalization.
- `enrichRowsWithPolicy` — given identity rows + a `PolicyManifest`, matching
  rows gain all five fields (with canonical Asuransi + normalized dates), and a
  non-matching row is untouched.
- Export integration — a row carrying the five insurance keys produces the
  correct `AA..AE` cells through the real `generateInlineSiskopatuhExcel`.
- `ProcessDocumentsSync` integration — fake identity analyzer + fake
  `PolicyExtractor`; a session of (passport image + policy "pdf") returns the
  jamaah row enriched with the fake manifest's insurance data, in order.

Live validation (during implementation, not a unit test): run the real sample
PDF text through the OpenCode extractor with the server key and confirm the
manifest JSON (22 entries, correct passports/policy numbers/dates), exactly as
the vision OCR was validated.

## Out of scope (v1)

- Vision/OCR extraction of **scanned** (image-only) policy PDFs. v1 supports
  digitally-generated text PDFs (the norm). A vision fallback can be added later
  behind the same `PolicyExtractor` interface.
- Cross-session enrichment (uploading the policy in a different session than the
  identity docs). The chosen workflow is single-session combined upload.
- Per-insurer layout tuning beyond what the general LLM prompt handles.
- BPJS column (not present on a travel-insurance policy).

## Global constraints

- Module `github.com/jamaah-in/v2`; Go; existing OpenCode provider + config.
- Insurance column **values** must conform to the Siskopatuh dropdowns where one
  exists (Asuransi → Sheet2 list); dates `yyyy-mm-dd`. Consistent with the
  2026-06-23 canonical-values calibration ([[opencode-zen-models]] context).
- No new dependency: `pdftotext` already ships in the ai-ocr-service container
  (poppler-utils 23.10.0).
- No "Co-Authored-By"/AI trailer in commits.

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jamaah-in/v2/internal/shared/config"
)

// PolicyEntry is one person's row from a POLIS manifest, keyed by passport. The
// manifest also carries the jamaah's name and birthdate, so a person listed in
// the policy but not yet scanned can still be seeded as a row.
type PolicyEntry struct {
	Nama         string
	NoIdentitas  string // passport / ID number (join key)
	TanggalLahir string // yyyy-mm-dd
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

// PolicyExtractor turns a POLIS PDF into a structured manifest. It gets both the
// raw bytes (to render the manifest table page as an image — accurate for names
// that wrap across lines) and the extracted text (to locate that page and read
// the cover-level Asuransi / issue date).
type PolicyExtractor interface {
	ExtractManifest(ctx context.Context, pdfData []byte, pdfText string) (*PolicyManifest, error)
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
		// input contains a full canonical name, or a long input is contained in
		// a canonical name (e.g. "ASKRIDA SYARIAH" ⊂ "ASURANSI ASKRIDA SYARIAH").
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

// parsePolicyJSON converts the LLM's manifest JSON into a PolicyManifest,
// normalising dates and dropping entries without an identity number.
func parsePolicyJSON(s string) (*PolicyManifest, error) {
	var raw struct {
		Asuransi     string `json:"asuransi"`
		TanggalInput string `json:"tanggal_input_polis"`
		Peserta      []struct {
			Nama         string `json:"nama"`
			NoIdentitas  string `json:"no_identitas"`
			TanggalLahir string `json:"tanggal_lahir"`
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
			Nama:         strings.TrimSpace(p.Nama),
			NoIdentitas:  strings.TrimSpace(p.NoIdentitas),
			TanggalLahir: normalizeDate(p.TanggalLahir),
			NoPolis:      strings.TrimSpace(p.NoPolis),
			TanggalAwal:  normalizeDate(p.TanggalAwal),
			TanggalAkhir: normalizeDate(p.TanggalAkhir),
		})
	}
	return m, nil
}

// rowPasporKey is a jamaah row's passport join key: no_paspor, or no_identitas
// when the row's identity type is a passport. Reuses fieldGetter for value reads.
func rowPasporKey(row map[string]any) string {
	g := fieldGetter{row}
	if p := g.first("no_paspor"); p != "" {
		return normPaspor(p)
	}
	if strings.EqualFold(g.first("jenis_identitas"), "PASPOR") {
		return normPaspor(g.first("no_identitas"))
	}
	return ""
}

const policyPrompt = `Anda mengekstrak data dari SERTIFIKAT/POLIS asuransi perjalanan umrah Indonesia (grup).
Dari teks dokumen, kembalikan HANYA JSON dengan bentuk:
{
  "asuransi": "<nama perusahaan asuransi / Pengelola>",
  "tanggal_input_polis": "<tanggal terbit sertifikat, format yyyy-mm-dd>",
  "peserta": [
    {"nama":"<nama lengkap peserta>", "no_identitas":"<nomor paspor/identitas>",
     "tanggal_lahir":"<tanggal lahir yyyy-mm-dd>", "no_polis":"<nomor polis peserta>",
     "tanggal_awal_polis":"<tanggal keberangkatan yyyy-mm-dd>", "tanggal_akhir_polis":"<tanggal kepulangan yyyy-mm-dd>"}
  ]
}
Aturan: ambil SEMUA baris peserta dari tabel MANIFEST (kolom NAMA, NO IDENTITAS, TANGGAL LAHIR, NO POLIS, TANGGAL KEBERANGKATAN, TANGGAL KEPULANGAN).
no_identitas = nomor paspor/identitas peserta. Semua tanggal format yyyy-mm-dd. Jika ragu suatu field, kosongkan. Tanpa teks lain selain JSON.`

const policyVisionPrompt = `Gambar berikut adalah tabel MANIFEST dari sertifikat/polis asuransi perjalanan umrah (grup).
Baca tabel dari GAMBAR — setiap baris adalah satu peserta. Cocokkan dengan teliti pada baris yang sama: kolom NAMA, NO IDENTITAS, TANGGAL LAHIR, NO POLIS, TANGGAL KEBERANGKATAN, TANGGAL KEPULANGAN. Nama bisa lebih dari satu kata atau menyambung ke baris berikutnya — ambil nama lengkapnya untuk paspor di baris itu.
Kembalikan HANYA JSON:
{
  "asuransi": "<nama perusahaan asuransi / Pengelola, dari teks konteks>",
  "tanggal_input_polis": "<tanggal terbit sertifikat yyyy-mm-dd, dari teks konteks>",
  "peserta": [
    {"nama":"<nama lengkap>", "no_identitas":"<nomor paspor/identitas>", "tanggal_lahir":"<yyyy-mm-dd>", "no_polis":"<nomor polis>", "tanggal_awal_polis":"<keberangkatan yyyy-mm-dd>", "tanggal_akhir_polis":"<kepulangan yyyy-mm-dd>"}
  ]
}
Ambil SEMUA baris peserta. Semua tanggal yyyy-mm-dd. Jika ragu suatu field, kosongkan. Tanpa teks lain selain JSON.`

// ExtractManifest reads the POLIS manifest. It renders the manifest table page(s)
// to image and uses vision (the table reads cleanly there, whereas pdftotext
// splits multi-line names and mis-aligns them with rows); the document text
// supplies the cover-level Asuransi + issue date. Falls back to a text-only pass
// when no manifest page can be rendered (e.g. a scanned policy).
func (a *OpenCodeAnalyzer) ExtractManifest(ctx context.Context, pdfData []byte, pdfText string) (*PolicyManifest, error) {
	if !a.Available() {
		return nil, fmt.Errorf("opencode analyzer not configured (OPENCODE_API_KEY missing)")
	}
	imgs := manifestPageImages(ctx, pdfData, pdfText)
	if len(imgs) == 0 {
		return a.extractManifestText(ctx, pdfText)
	}

	docText := pdfText
	if len(docText) > 8000 {
		docText = docText[:8000]
	}
	content := []map[string]any{{"type": "text", "text": policyVisionPrompt}}
	for _, img := range imgs {
		content = append(content, map[string]any{
			"type": "image_url",
			"image_url": map[string]any{
				"url": "data:image/png;base64," + base64.StdEncoding.EncodeToString(img),
			},
		})
	}
	content = append(content, map[string]any{
		"type": "text",
		"text": "Teks halaman sampul & catatan (untuk 'asuransi' dan 'tanggal_input_polis'):\n" + docText,
	})
	out, err := a.chat(ctx, content, 4096)
	if err != nil {
		return nil, err
	}
	return parsePolicyJSON(out)
}

// extractManifestText is the text-only fallback when the manifest page cannot be
// rendered to an image.
func (a *OpenCodeAnalyzer) extractManifestText(ctx context.Context, pdfText string) (*PolicyManifest, error) {
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

// manifestPageImages locates the manifest table page(s) from the per-page text
// (pdftotext separates pages with form feeds) and renders them to PNG, capped at
// three pages.
func manifestPageImages(ctx context.Context, data []byte, text string) [][]byte {
	var imgs [][]byte
	for i, pg := range strings.Split(text, "\f") {
		u := strings.ToUpper(pg)
		isManifest := strings.Contains(u, "MANIFEST") ||
			(strings.Contains(u, "NO POLIS") &&
				(strings.Contains(u, "KEBERANGKATAN") || strings.Contains(u, "IDENTITAS")))
		if !isManifest {
			continue
		}
		if img, err := rasterizePDFPage(ctx, data, i+1); err == nil && len(img) > 0 {
			imgs = append(imgs, img)
			if len(imgs) >= 3 {
				break
			}
		}
	}
	return imgs
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

// existingPasporKeys is the set of passport join keys already present in the
// scanned rows, so the policy lane only seeds rows for jamaah not yet scanned.
func existingPasporKeys(data []any) map[string]bool {
	set := map[string]bool{}
	for _, item := range data {
		if m, ok := item.(map[string]any); ok {
			if k := rowPasporKey(m); k != "" {
				set[k] = true
			}
		}
	}
	return set
}

// policyEntryToRow builds a jamaah row from a manifest entry alone — used when a
// person listed in the POLIS has no scanned identity document. It carries the
// name, passport, birthdate and the full insurance block; identity-only fields
// (address, gender, Title, …) stay empty until a passport/KTP is scanned.
func policyEntryToRow(e PolicyEntry, asuransi, tglInput string) map[string]any {
	row := map[string]any{
		"source_doc_type":    "polis",
		"siskopatuh_version": "2.0",
	}
	if e.Nama != "" {
		row["nama"] = e.Nama
		row["nama_paspor"] = e.Nama
	}
	if e.NoIdentitas != "" {
		row["no_paspor"] = e.NoIdentitas
		row["no_identitas"] = e.NoIdentitas
		row["jenis_identitas"] = "PASPOR"
	}
	if e.TanggalLahir != "" {
		row["tanggal_lahir"] = e.TanggalLahir
	}
	if a := mapAsuransi(asuransi); a != "" {
		row["asuransi"] = a
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
	return row
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

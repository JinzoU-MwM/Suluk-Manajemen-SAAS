package service

import (
	"context"
	"encoding/json"
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

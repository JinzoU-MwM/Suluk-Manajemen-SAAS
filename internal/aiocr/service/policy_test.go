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
		"17-June-2026":  "2026-06-17",
		"01-Jul-2026":   "2026-07-01",
		"2026-07-09":    "2026-07-09",
		"9/7/2026":      "2026-07-09",
		"":              "",
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

func TestParsePolicyJSON(t *testing.T) {
	in := `{"asuransi":"PT. ASURANSI ASKRIDA SYARIAH","tanggal_input_polis":"17-June-2026",
	"peserta":[{"nama":"LESTARI EKA CITRA","no_identitas":"X2664222","tanggal_lahir":"15-Jun-1987","no_polis":"122015022600316-000043","tanggal_awal_polis":"01-Jul-2026","tanggal_akhir_polis":"09-Jul-2026"},
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
	if e.Nama != "LESTARI EKA CITRA" || e.NoIdentitas != "X2664222" || e.TanggalLahir != "1987-06-15" ||
		e.NoPolis != "122015022600316-000043" || e.TanggalAwal != "2026-07-01" || e.TanggalAkhir != "2026-07-09" {
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

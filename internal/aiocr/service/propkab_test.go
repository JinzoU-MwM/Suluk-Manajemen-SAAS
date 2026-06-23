package service

import "testing"

func TestMapProvinsi(t *testing.T) {
	cases := map[string]string{
		"JAWA BARAT":          "JAWA BARAT",
		"Jawa Barat":          "JAWA BARAT",
		"PROVINSI JAWA BARAT": "JAWA BARAT",
		"BALI":                "B A L I", // template stores decorative spacing
		"Bali":                "B A L I",
		"RIAU":                "R I A U",
		"JAMBI":               "J A M B I",
		"DKI JAKARTA":         "DKI JAKARTA",
		"JAKARTA":             "DKI JAKARTA", // alias
		"YOGYAKARTA":          "D.I. YOGYAKARTA",
		"DI YOGYAKARTA":       "D.I. YOGYAKARTA",
		"":                    "",
	}
	for in, want := range cases {
		if got := mapProvinsi(in); got != want {
			t.Errorf("mapProvinsi(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapKabupaten(t *testing.T) {
	cases := []struct{ prov, kab, want string }{
		{"JAWA BARAT", "KOTA BANDUNG", "KOTA BANDUNG"},
		{"JAWA BARAT", "Kota Bandung", "KOTA BANDUNG"},
		{"JAWA BARAT", "KABUPATEN BANDUNG", "KAB. BANDUNG"},
		{"JAWA BARAT", "KAB. BANDUNG", "KAB. BANDUNG"}, // idempotent
		{"D.I. YOGYAKARTA", "BANTUL", "KAB. B A N T U L"},
		{"D.I. YOGYAKARTA", "KABUPATEN BANTUL", "KAB. B A N T U L"},
		// province blank -> global fallback still resolves a unique kabupaten
		{"", "KOTA BANDUNG", "KOTA BANDUNG"},
		{"JAWA BARAT", "", ""},
	}
	for _, c := range cases {
		if got := mapKabupaten(c.prov, c.kab); got != c.want {
			t.Errorf("mapKabupaten(%q,%q) = %q, want %q", c.prov, c.kab, got, c.want)
		}
	}
}

// TestProvinsiKabupatenInReference guards that canonical outputs are real
// members of the PropKab reference data.
func TestProvinsiKabupatenInReference(t *testing.T) {
	provSet := map[string]bool{}
	for _, p := range siskopatuhProvinces {
		provSet[p] = true
	}
	if got := mapProvinsi("JAWA TENGAH"); !provSet[got] {
		t.Errorf("mapProvinsi gave %q, not in reference province list", got)
	}
	jb := siskopatuhKabupaten["JAWA BARAT"]
	kabSet := map[string]bool{}
	for _, k := range jb {
		kabSet[k] = true
	}
	if got := mapKabupaten("JAWA BARAT", "KOTA BANDUNG"); !kabSet[got] {
		t.Errorf("mapKabupaten gave %q, not in JAWA BARAT reference list", got)
	}
}

package service

import "testing"

// Canonical allowed-value sets copied verbatim from template jamaah.xlsm Sheet2.
// Every mapper output (when non-empty) MUST be a member of its set, or Siskopatuh
// rejects the uploaded cell.
var (
	allowedJenisIdentitas  = map[string]bool{"NIK": true, "KITAS": true, "KITAP": true, "PASPOR": true}
	allowedStatusNikah     = map[string]bool{"BELUM MENIKAH": true, "MENIKAH": true, "JANDA / DUDA": true}
	allowedKewarganegaraan = map[string]bool{"WNI": true, "WNA": true}
	allowedTitle           = map[string]bool{"TUAN": true, "NONA": true, "NYONYA": true}
	allowedPendidikan      = map[string]bool{
		"TIDAK SEKOLAH": true, "SD/MI": true, "SMP/MTS": true, "SMA/MA": true,
		"D1": true, "D2": true, "D3": true, "D4/S1": true, "S2": true, "S3": true,
	}
	allowedPekerjaan = map[string]bool{
		"PNS": true, "PEG. SWASTA": true, "WIRAUSAHA": true, "TNI / POLRI": true,
		"PETANI": true, "NELAYAN": true, "LAINNYA": true, "TIDAK BEKERJA": true,
	}
)

func TestMapStatusNikah(t *testing.T) {
	cases := map[string]string{
		"KAWIN":         "MENIKAH",
		"MENIKAH":       "MENIKAH", // idempotent
		"BELUM KAWIN":   "BELUM MENIKAH",
		"BELUM MENIKAH": "BELUM MENIKAH", // idempotent
		"CERAI HIDUP":   "JANDA / DUDA",
		"CERAI MATI":    "JANDA / DUDA",
		"JANDA / DUDA":  "JANDA / DUDA", // idempotent
		"JANDA":         "JANDA / DUDA",
		"DUDA":          "JANDA / DUDA",
		"":              "",
	}
	for in, want := range cases {
		if got := mapStatusNikah(in); got != want {
			t.Errorf("mapStatusNikah(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormJenisIdentitas(t *testing.T) {
	cases := []struct {
		existing, nik, paspor, want string
	}{
		{"", "3273123456780001", "", "NIK"}, // KTP, derived
		{"", "", "X2664222", "PASPOR"},      // passport, derived
		{"KTP", "3273...", "", "NIK"},       // legacy "KTP" normalized
		{"Paspor", "", "E123", "PASPOR"},    // legacy "Paspor" normalized
		{"PASPOR", "", "E123", "PASPOR"},    // idempotent
		{"NIK", "327...", "", "NIK"},        // idempotent
		{"KITAS", "", "", "KITAS"},
		{"KITAP", "", "", "KITAP"},
		{"", "", "", ""},
	}
	for _, c := range cases {
		if got := normJenisIdentitas(c.existing, c.nik, c.paspor); got != c.want {
			t.Errorf("normJenisIdentitas(%q,%q,%q) = %q, want %q", c.existing, c.nik, c.paspor, got, c.want)
		}
	}
}

func TestMapKewarganegaraan(t *testing.T) {
	cases := map[string]string{
		"WNI": "WNI", "INDONESIA": "WNI", "INDONESIAN": "WNI",
		"WNA": "WNA", "MALAYSIA": "WNA", "": "",
	}
	for in, want := range cases {
		if got := mapKewarganegaraan(in); got != want {
			t.Errorf("mapKewarganegaraan(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapTitle(t *testing.T) {
	cases := []struct{ gender, status, want string }{
		{"LAKI-LAKI", "KAWIN", "TUAN"},
		{"PRIA", "", "TUAN"},
		{"PEREMPUAN", "KAWIN", "NYONYA"},
		{"PEREMPUAN", "BELUM KAWIN", "NONA"},
		{"PEREMPUAN", "CERAI MATI", "NYONYA"}, // widow keeps NYONYA
		{"PEREMPUAN", "", "NONA"},             // unknown status -> NONA
		{"", "", ""},
	}
	for _, c := range cases {
		if got := mapTitle(c.gender, c.status); got != c.want {
			t.Errorf("mapTitle(%q,%q) = %q, want %q", c.gender, c.status, got, c.want)
		}
	}
}

func TestMapPendidikan(t *testing.T) {
	cases := map[string]string{
		"SD": "SD/MI", "SD/MI": "SD/MI",
		"SMP": "SMP/MTS", "SLTP": "SMP/MTS",
		"SMA": "SMA/MA", "SMK": "SMA/MA", "SLTA": "SMA/MA", "MA": "SMA/MA",
		"D3": "D3", "DIPLOMA III": "D3",
		"S1": "D4/S1", "SARJANA": "D4/S1", "D4": "D4/S1",
		"S2": "S2", "MAGISTER": "S2",
		"S3": "S3", "DOKTOR": "S3",
		"TIDAK SEKOLAH": "TIDAK SEKOLAH",
		"":              "",
	}
	for in, want := range cases {
		if got := mapPendidikan(in); got != want {
			t.Errorf("mapPendidikan(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapPekerjaan(t *testing.T) {
	cases := map[string]string{
		"PEGAWAI NEGERI SIPIL": "PNS", "PNS": "PNS",
		"KARYAWAN SWASTA": "PEG. SWASTA", "PEGAWAI SWASTA": "PEG. SWASTA",
		"BURUH": "PEG. SWASTA", "GURU": "PEG. SWASTA",
		"WIRASWASTA": "WIRAUSAHA", "WIRAUSAHA": "WIRAUSAHA", "PEDAGANG": "WIRAUSAHA",
		"TNI": "TNI / POLRI", "POLRI": "TNI / POLRI", "KEPOLISIAN": "TNI / POLRI",
		"PETANI": "PETANI", "NELAYAN": "NELAYAN",
		"MENGURUS RUMAH TANGGA": "TIDAK BEKERJA", "PELAJAR/MAHASISWA": "TIDAK BEKERJA",
		"BELUM/TIDAK BEKERJA": "TIDAK BEKERJA",
		"SENIMAN":             "LAINNYA", "DOKTER GIGI": "PEG. SWASTA", // doctor -> employee bucket
		"": "",
	}
	for in, want := range cases {
		if got := mapPekerjaan(in); got != want {
			t.Errorf("mapPekerjaan(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapProviderVisa(t *testing.T) {
	cases := map[string]string{
		"Saudi Digital Embassy":    "B2C",
		"SAUDI DIGITAL EMBASSY":    "B2C",
		"B2C":                      "B2C",
		"E-Visa":                   "B2C",
		"e visa elektronik":        "B2C",
		"Nusuk":                    "B2C",
		"":                         "",
		"PT. AERO GLOBE INDONESIA": "PT. AERO GLOBE INDONESIA",
	}
	for in, want := range cases {
		if got := mapProviderVisa(in); got != want {
			t.Errorf("mapProviderVisa(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestMappersStayInAllowedSets guards against any mapper ever emitting a value
// outside its Sheet2 dropdown set (empty is always allowed = operator fills in).
func TestMappersStayInAllowedSets(t *testing.T) {
	probe := []string{"", "x", "UNKNOWN", "KAWIN", "WNI", "SMA", "KARYAWAN", "PEREMPUAN", "laki-laki", "Paspor"}
	check := func(name string, got string, set map[string]bool) {
		if got != "" && !set[got] {
			t.Errorf("%s produced %q which is not an allowed Siskopatuh value", name, got)
		}
	}
	for _, p := range probe {
		check("mapStatusNikah", mapStatusNikah(p), allowedStatusNikah)
		check("mapKewarganegaraan", mapKewarganegaraan(p), allowedKewarganegaraan)
		check("mapPendidikan", mapPendidikan(p), allowedPendidikan)
		check("mapPekerjaan", mapPekerjaan(p), allowedPekerjaan)
		check("mapTitle", mapTitle(p, "KAWIN"), allowedTitle)
		check("normJenisIdentitas", normJenisIdentitas(p, "", ""), allowedJenisIdentitas)
	}
}

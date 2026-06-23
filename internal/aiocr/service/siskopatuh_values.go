package service

import "strings"

// This file maps free-text OCR output to the canonical Siskopatuh dropdown
// values defined in template jamaah.xlsm (Sheet2). Siskopatuh validates each
// uploaded cell against these exact lists, so any other value is rejected.
// Every mapper is idempotent on already-canonical input and returns "" when it
// cannot confidently classify (an empty cell is left for the operator to fill).

// normUpper trims, uppercases, and collapses internal runs of whitespace to a
// single space so comparisons ignore OCR spacing noise.
func normUpper(s string) string {
	return strings.Join(strings.Fields(strings.ToUpper(s)), " ")
}

// mapTitle derives the Title/Gelar (Sheet2 col H): TUAN / NONA / NYONYA.
// Male -> TUAN. Female -> NYONYA when ever-married (married or widowed),
// otherwise NONA.
func mapTitle(gender, status string) string {
	g := normUpper(gender)
	switch {
	case g == "":
		return ""
	case strings.Contains(g, "PEREMPUAN") || strings.Contains(g, "WANITA") || strings.Contains(g, "FEMALE") || g == "P":
		st := mapStatusNikah(status)
		if st == "MENIKAH" || st == "JANDA / DUDA" {
			return "NYONYA"
		}
		return "NONA"
	case strings.Contains(g, "LAKI") || strings.Contains(g, "PRIA") || strings.Contains(g, "MALE") || g == "L":
		return "TUAN"
	default:
		return ""
	}
}

// mapStatusNikah collapses any marital-status text to the Sheet2 col C values:
// BELUM MENIKAH / MENIKAH / JANDA / DUDA. Unknown/empty -> "".
func mapStatusNikah(s string) string {
	t := normUpper(s)
	switch {
	case t == "":
		return ""
	case strings.Contains(t, "BELUM") || strings.Contains(t, "TIDAK"):
		return "BELUM MENIKAH" // BELUM KAWIN / TIDAK KAWIN
	case strings.Contains(t, "CERAI") || strings.Contains(t, "JANDA") || strings.Contains(t, "DUDA"):
		return "JANDA / DUDA"
	case strings.Contains(t, "KAWIN") || strings.Contains(t, "NIKAH"):
		return "MENIKAH"
	default:
		return "" // don't fabricate a marital status
	}
}

// normJenisIdentitas resolves the Jenis Identitas (Sheet2 col E):
// NIK / KITAS / KITAP / PASPOR. It honours an explicit existing value (and
// normalises legacy "KTP"/"Paspor"), otherwise derives from which number is
// present (a passport number => PASPOR, an NIK => NIK).
func normJenisIdentitas(existing, nik, noPaspor string) string {
	e := normUpper(existing)
	switch {
	case strings.Contains(e, "PASPOR") || strings.Contains(e, "PASSPORT"):
		return "PASPOR"
	case strings.Contains(e, "KITAS"):
		return "KITAS"
	case strings.Contains(e, "KITAP"):
		return "KITAP"
	case e == "NIK" || strings.Contains(e, "KTP"):
		return "NIK"
	}
	if strings.TrimSpace(noPaspor) != "" {
		return "PASPOR"
	}
	if strings.TrimSpace(nik) != "" {
		return "NIK"
	}
	return ""
}

// mapKewarganegaraan -> Sheet2 col D: WNI / WNA. Empty -> "".
func mapKewarganegaraan(s string) string {
	t := normUpper(s)
	switch {
	case t == "":
		return ""
	case strings.Contains(t, "WNI") || strings.Contains(t, "INDONESIA") || t == "ID" || t == "IDN":
		return "WNI"
	default:
		return "WNA" // any explicit non-Indonesian nationality (incl. "WNA")
	}
}

// mapPendidikan -> Sheet2 col B (last education). Education is not on a KTP/
// passport, so this is usually empty; it matters for Kartu Keluarga scans.
// Unknown -> "".
func mapPendidikan(s string) string {
	t := normUpper(s)
	switch {
	case t == "":
		return ""
	case strings.Contains(t, "TIDAK") || strings.Contains(t, "BELUM SEKOLAH"):
		return "TIDAK SEKOLAH"
	case strings.Contains(t, "S3") || strings.Contains(t, "DOKTOR") || strings.Contains(t, "PHD") || strings.Contains(t, "DOCTORAL"):
		return "S3"
	case strings.Contains(t, "S2") || strings.Contains(t, "MAGISTER") || strings.Contains(t, "MASTER"):
		return "S2"
	case strings.Contains(t, "D4") || strings.Contains(t, "S1") || strings.Contains(t, "SARJANA") ||
		strings.Contains(t, "DIPLOMA 4") || strings.Contains(t, "DIPLOMA IV") || strings.Contains(t, "D IV"):
		return "D4/S1"
	case strings.Contains(t, "D3") || strings.Contains(t, "DIPLOMA 3") || strings.Contains(t, "DIPLOMA III") || strings.Contains(t, "D III"):
		return "D3"
	case strings.Contains(t, "D2") || strings.Contains(t, "DIPLOMA 2") || strings.Contains(t, "DIPLOMA II"):
		return "D2"
	case strings.Contains(t, "D1") || strings.Contains(t, "DIPLOMA 1") || strings.Contains(t, "DIPLOMA I"):
		return "D1"
	case strings.Contains(t, "SMA") || strings.Contains(t, "SMK") || strings.Contains(t, "SLTA") ||
		strings.Contains(t, "ALIYAH") || t == "MA":
		return "SMA/MA"
	case strings.Contains(t, "SMP") || strings.Contains(t, "MTS") || strings.Contains(t, "SLTP") || strings.Contains(t, "TSANAWIYAH"):
		return "SMP/MTS"
	case strings.Contains(t, "SD") || strings.Contains(t, "IBTIDAIYAH") || t == "MI":
		return "SD/MI"
	default:
		return ""
	}
}

// mapPekerjaan -> Sheet2 col A, one of 8 buckets. Any recognised-but-unbucketed
// job falls to LAINNYA (the template's explicit catch-all). Empty -> "".
// Order matters: more specific prefixes are tested before broader ones
// (PEGAWAI NEGERI before PEGAWAI; WIRASWASTA before SWASTA).
func mapPekerjaan(s string) string {
	t := normUpper(s)
	switch {
	case t == "":
		return ""
	case strings.Contains(t, "TIDAK BEKERJA") || strings.Contains(t, "BELUM") ||
		strings.Contains(t, "PELAJAR") || strings.Contains(t, "MAHASISWA") ||
		strings.Contains(t, "MENGURUS RUMAH") || strings.Contains(t, "IBU RUMAH") ||
		strings.Contains(t, "IRT") || strings.Contains(t, "PENSIUN"):
		return "TIDAK BEKERJA"
	case strings.Contains(t, "PNS") || strings.Contains(t, "PEGAWAI NEGERI") ||
		strings.Contains(t, "APARATUR") || t == "ASN":
		return "PNS"
	case strings.Contains(t, "TNI") || strings.Contains(t, "POLRI") || strings.Contains(t, "POLISI") ||
		strings.Contains(t, "TENTARA") || strings.Contains(t, "KEPOLISIAN"):
		return "TNI / POLRI"
	case strings.Contains(t, "NELAYAN"):
		return "NELAYAN"
	case strings.Contains(t, "PETANI") || strings.Contains(t, "PEKEBUN") ||
		strings.Contains(t, "BERKEBUN") || strings.Contains(t, "PETERNAK"):
		return "PETANI"
	case strings.Contains(t, "WIRA") || strings.Contains(t, "PEDAGANG") ||
		strings.Contains(t, "DAGANG") || strings.Contains(t, "PENGUSAHA") || strings.Contains(t, "USAHA"):
		return "WIRAUSAHA"
	case strings.Contains(t, "SWASTA") || strings.Contains(t, "KARYAWAN") || strings.Contains(t, "KARYAWATI") ||
		strings.Contains(t, "PEGAWAI") || strings.Contains(t, "BURUH") ||
		strings.Contains(t, "BUMN") || strings.Contains(t, "BUMD") ||
		strings.Contains(t, "GURU") || strings.Contains(t, "DOSEN") ||
		strings.Contains(t, "DOKTER") || strings.Contains(t, "PERAWAT") || strings.Contains(t, "BIDAN"):
		return "PEG. SWASTA"
	default:
		return "LAINNYA"
	}
}

package service

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestExportMatchesJamaahTemplate(t *testing.T) {
	records := []map[string]any{
		{ // passport, female, married
			"nama": "LESTARI EKA CITRA", "nama_paspor": "LESTARI EKA CITRA",
			"no_paspor": "X2664222", "gender": "PEREMPUAN", "status_pernikahan": "KAWIN",
			"tanggal_lahir": "1987-06-15", "kewarganegaraan": "WNI",
		},
		{ // KTP, male, single (jamaah-member shape: nik / jenis_kelamin / status_perkawinan)
			"nama": "BUDI SANTOSO", "nik": "3273123456780001",
			"jenis_kelamin": "LAKI-LAKI", "status_perkawinan": "BELUM KAWIN",
		},
	}
	data, err := generateInlineSiskopatuhExcel(records)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	cell := func(c string) string { v, _ := f.GetCellValue("Sheet1", c); return v }

	// Header row must match the template verbatim.
	for c, want := range map[string]string{
		"A1": "Title", "B1": "Nama (Sesuai Dengan nama Pada Kartu Vaksin)",
		"D1": "Jenis Identitas", "E1": "No Identitas", "G1": "No Paspor",
		"T1": "Status Pernikahan", "AF1": "No BPJS",
	} {
		if got := cell(c); got != want {
			t.Errorf("header %s = %q, want %q", c, got, want)
		}
	}

	// Row 2: passport, female, married.
	for c, want := range map[string]string{
		"A2": "NYONYA", "D2": "PASPOR", "E2": "X2664222", "G2": "X2664222", "T2": "MENIKAH",
	} {
		if got := cell(c); got != want {
			t.Errorf("passport row %s = %q, want %q", c, got, want)
		}
	}

	// Row 3: KTP, male, single.
	for c, want := range map[string]string{
		"A3": "TUAN", "D3": "NIK", "E3": "3273123456780001", "T3": "BELUM MENIKAH",
	} {
		if got := cell(c); got != want {
			t.Errorf("ktp row %s = %q, want %q", c, got, want)
		}
	}
	// Insurance columns stay blank (not from OCR).
	if got := cell("AA2"); got != "" {
		t.Errorf("Asuransi should be blank, got %q", got)
	}

	// Nama (col B) and Nama Paspor (col F) must be identical in every row — the
	// passport row carries both, the KTP row only "nama" (F used to come out blank).
	if b, f := cell("B2"), cell("F2"); b == "" || b != f {
		t.Errorf("row2: Nama B2=%q must equal Nama Paspor F2=%q", b, f)
	}
	if b, f := cell("B3"), cell("F3"); b == "" || b != f {
		t.Errorf("row3: Nama B3=%q must equal Nama Paspor F3=%q", b, f)
	}
}

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

func TestSanitizeCellValue(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"Ahmad Yani", "Ahmad Yani"},
		{"=cmd|'/c calc.exe'!A1", "'=cmd|'/c calc.exe'!A1"},
		{"+62812345678", "'+62812345678"},
		{"-1234", "'-1234"},
		{"@SUM(A1:A9)", "'@SUM(A1:A9)"},
		{"Jl. Sudirman No.1-2", "Jl. Sudirman No.1-2"},
	}
	for _, c := range cases {
		if got := sanitizeCellValue(c.in); got != c.want {
			t.Errorf("sanitizeCellValue(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestExportNeutralizesFormulaInjection is the AIOCR-2 regression: an OCR
// result crafted to look like a spreadsheet formula must come out of a real
// generated .xlsx as literal text, not something Excel would execute.
func TestExportNeutralizesFormulaInjection(t *testing.T) {
	records := []map[string]any{{
		"nama":      "=cmd|'/c calc.exe'!A1",
		"no_paspor": "X1234567",
		"alamat":    "+62812345678 injected",
	}}
	data, err := generateInlineSiskopatuhExcel(records)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	cell := func(c string) string { v, _ := f.GetCellValue("Sheet1", c); return v }

	if got := cell("B2"); got != "'=cmd|'/c calc.exe'!A1" {
		t.Errorf("Nama B2 = %q, want a literal-text-escaped formula", got)
	}
	if got := cell("L2"); got != "'+62812345678 injected" {
		t.Errorf("Alamat L2 = %q, want a literal-text-escaped leading '+'", got)
	}
}

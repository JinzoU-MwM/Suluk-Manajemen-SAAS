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
		"A2": "NYONYA", "D2": "Paspor", "E2": "X2664222", "G2": "X2664222", "T2": "MENIKAH",
	} {
		if got := cell(c); got != want {
			t.Errorf("passport row %s = %q, want %q", c, got, want)
		}
	}

	// Row 3: KTP, male, single.
	for c, want := range map[string]string{
		"A3": "TUAN", "D3": "KTP", "E3": "3273123456780001", "T3": "BELUM MENIKAH",
	} {
		if got := cell(c); got != want {
			t.Errorf("ktp row %s = %q, want %q", c, got, want)
		}
	}
	// Insurance columns stay blank (not from OCR).
	if got := cell("AA2"); got != "" {
		t.Errorf("Asuransi should be blank, got %q", got)
	}
}

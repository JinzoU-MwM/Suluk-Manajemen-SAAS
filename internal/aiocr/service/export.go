package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jamaah-in/v2/internal/aiocr/model"

	"github.com/xuri/excelize/v2"
)

// templateColumn is one column of the "template jamaah.xlsm" Siskopatuh upload
// format. header is the EXACT spreadsheet header (must match the government
// template verbatim — including spacing/casing); value derives the cell from a
// scanned record.
type templateColumn struct {
	header string
	value  func(g fieldGetter) string
}

// templateColumns mirrors template jamaah.xlsm (Sheet1) header text + order
// EXACTLY. Do not reword headers — Siskopatuh matches columns by header.
var templateColumns = []templateColumn{
	{"Title", func(g fieldGetter) string {
		return mapTitle(g.first("gender", "jenis_kelamin"), g.first("status_pernikahan", "status_perkawinan"))
	}},
	{"Nama (Sesuai Dengan nama Pada Kartu Vaksin)", func(g fieldGetter) string { return g.first("nama", "nama_paspor") }},
	{"Nama Ayah", func(g fieldGetter) string { return g.first("nama_ayah") }},
	{"Jenis Identitas", func(g fieldGetter) string {
		return normJenisIdentitas(g.first("jenis_identitas"), g.first("nik"), g.first("no_paspor"))
	}},
	{"No Identitas", func(g fieldGetter) string { return g.first("no_paspor", "no_identitas", "nik") }},
	// Nama Paspor must equal the Nama column — the template uses the same name for
	// both. Resolve with the identical expression so they are always byte-identical.
	{"Nama Paspor", func(g fieldGetter) string { return g.first("nama", "nama_paspor") }},
	{"No Paspor", func(g fieldGetter) string { return g.first("no_paspor") }},
	{"Tanggal Dikeluarkan Paspor(yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_paspor", "tanggal_terbit_paspor") }},
	{"Kota Paspor", func(g fieldGetter) string { return g.first("kota_paspor") }},
	{"Tempat Lahir", func(g fieldGetter) string { return g.first("tempat_lahir") }},
	{"Tanggal Lahir(yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_lahir") }},
	{"Alamat", func(g fieldGetter) string { return g.first("alamat") }},
	{"Provinsi", func(g fieldGetter) string { return mapProvinsi(g.first("provinsi")) }},
	{"Kabupaten", func(g fieldGetter) string {
		return mapKabupaten(mapProvinsi(g.first("provinsi")), g.first("kabupaten"))
	}},
	{"Kecamatan", func(g fieldGetter) string { return g.first("kecamatan") }},
	{"Kelurahan", func(g fieldGetter) string { return g.first("kelurahan") }},
	{"No. Telepon", func(g fieldGetter) string { return g.first("no_telepon") }},
	{"No Hp", func(g fieldGetter) string { return g.first("no_hp") }},
	{"KewargaNegaraan", func(g fieldGetter) string { return mapKewarganegaraan(g.first("kewarganegaraan")) }},
	{"Status Pernikahan", func(g fieldGetter) string { return mapStatusNikah(g.first("status_pernikahan", "status_perkawinan")) }},
	{"Pendidikan", func(g fieldGetter) string { return mapPendidikan(g.first("pendidikan")) }},
	{"Pekerjaan", func(g fieldGetter) string { return mapPekerjaan(g.first("pekerjaan")) }},
	{"Provider Visa", func(g fieldGetter) string { return mapProviderVisa(g.first("provider_visa")) }},
	{"No Visa", func(g fieldGetter) string { return g.first("no_visa") }},
	{"Tanggal Berlaku Visa (yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_visa") }},
	{"Tanggal Akhir  Visa (yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_visa_akhir") }},
	// Insurance / BPJS columns are not on identity documents — left blank for
	// the operator to fill in.
	{"Asuransi", func(g fieldGetter) string { return mapAsuransi(g.first("asuransi")) }},
	{"No Polis", func(g fieldGetter) string { return g.first("no_polis") }},
	{"Tanggal Input Polis (yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_input_polis") }},
	{"Tanggal Awal Polis (yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_awal_polis") }},
	{"Tanggal Akhir Polis (yyyy-mm-dd)", func(g fieldGetter) string { return g.first("tanggal_akhir_polis") }},
	{"No BPJS", func(g fieldGetter) string { return "" }},
}

// fieldGetter reads string fields from a record map, trying key aliases in order
// (records come in either the normalized OCR shape or the jamaah-member shape).
type fieldGetter struct{ m map[string]any }

func (g fieldGetter) first(keys ...string) string {
	for _, k := range keys {
		v, ok := g.m[k]
		if !ok || v == nil {
			continue
		}
		if s := strings.TrimSpace(fmt.Sprintf("%v", v)); s != "" {
			return s
		}
	}
	return ""
}

// writeSiskopatuhTemplate writes the records as the template jamaah.xlsm format.
func writeSiskopatuhTemplate(rows []fieldGetter) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0) // keep the default "Sheet1" — the gov template uses Sheet1

	for i, c := range templateColumns {
		_ = f.SetCellValue(sheet, fmt.Sprintf("%s1", columnLetter(i+1)), c.header)
	}
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#2563EB"}},
	})
	_ = f.SetCellStyle(sheet, "A1", fmt.Sprintf("%s1", columnLetter(len(templateColumns))), style)

	for r, g := range rows {
		for i, c := range templateColumns {
			if v := c.value(g); v != "" {
				_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", columnLetter(i+1), r+2), v)
			}
		}
	}
	for i := range templateColumns {
		col := columnLetter(i + 1)
		_ = f.SetColWidth(sheet, col, col, 22)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}

func generateSiskopatuhExcel(results []model.ScanResult) ([]byte, error) {
	rows := make([]fieldGetter, len(results))
	for i, sr := range results {
		rows[i] = fieldGetter{scanResultMap(sr)}
	}
	return writeSiskopatuhTemplate(rows)
}

func generateInlineSiskopatuhExcel(records []map[string]any) ([]byte, error) {
	rows := make([]fieldGetter, len(records))
	for i, rec := range records {
		rows[i] = fieldGetter{rec}
	}
	return writeSiskopatuhTemplate(rows)
}

// scanResultMap flattens a scan result's extracted + normalized data into one
// map the template columns can read (normalized values augment the raw ones).
func scanResultMap(sr model.ScanResult) map[string]any {
	m := map[string]any{}
	merge := func(v any) {
		if v == nil {
			return
		}
		if mm, ok := v.(map[string]any); ok {
			for k, val := range mm {
				m[k] = val
			}
			return
		}
		b, err := json.Marshal(v)
		if err != nil {
			return
		}
		var mm map[string]any
		if json.Unmarshal(b, &mm) == nil {
			for k, val := range mm {
				m[k] = val
			}
		}
	}
	merge(sr.ExtractedData)
	merge(sr.NormalizedData)
	return m
}

// ExportRecordsExcel builds a template jamaah.xlsm-format .xlsx from inline
// records (the scanner preview rows or group members the UI sends to
// /generate-excel).
func (s *AIOCRService) ExportRecordsExcel(records []map[string]any) ([]byte, error) {
	return generateInlineSiskopatuhExcel(records)
}

func columnLetter(n int) string {
	result := ""
	for n > 0 {
		n--
		result = string(rune('A'+n%26)) + result
		n /= 26
	}
	return result
}

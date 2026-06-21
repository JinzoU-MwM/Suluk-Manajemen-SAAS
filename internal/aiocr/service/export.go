package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jamaah-in/v2/internal/aiocr/model"

	"github.com/xuri/excelize/v2"
)

var siskopatuhColumns = []string{
	"No",
	"NAMA LENGKAP",
	"JENIS KELAMIN",
	"TEMPAT LAHIR",
	"TANGGAL LAHIR",
	"ALAMAT",
	"PROVINSI",
	"KABUPATEN / KOTA",
	"KECAMATAN",
	"KELURAHAN",
	"NO KTP / NIK",
	"NO PASPOR",
	"TANGGAL TERBIT PASPOR",
	"TANGGAL EXPIRED PASPOR",
	"TEMPAT TERBIT PASPOR",
	"NO TELEPON",
	"NO HP",
	"EMAIL",
	"KEWARGANEGARAAN",
	"STATUS PERKAWINAN",
	"PENDIDIKAN TERAKHIR",
	"PEKERJAAN",
	"GOLONGAN DARAH",
	"NAMA AYAH",
	"NO VISA",
	"TANGGAL TERBIT VISA",
	"TANGGAL EXPIRED VISA",
	"PROVIDER VISA",
	"UKURAN IHRAM / MUKENA",
	"UKURAN BAJU",
	"KONTAK DARURAT - NAMA",
	"KONTAK DARURAT - TELEPON",
}

func generateSiskopatuhExcel(results []model.ScanResult) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Siskopatuh"
	_ = f.SetSheetName("Sheet1", sheetName)

	for i, colName := range siskopatuhColumns {
		cell := fmt.Sprintf("%s1", columnLetter(i+1))
		_ = f.SetCellValue(sheetName, cell, colName)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#2563EB"}},
	})
	_ = f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", columnLetter(len(siskopatuhColumns))), style)

	for i, sr := range results {
		row := i + 2
		data := extractSiskopatuhRow(sr)
		data["No"] = fmt.Sprintf("%d", i+1) // fill row number

		for j, colName := range siskopatuhColumns {
			cell := fmt.Sprintf("%s%d", columnLetter(j+1), row)
			val := data[colName]
			if val != "" {
				_ = f.SetCellValue(sheetName, cell, val)
			}
		}
	}

	for i := 1; i <= len(siskopatuhColumns); i++ {
		col := columnLetter(i)
		_ = f.SetColWidth(sheetName, col, col, 25)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}

func extractSiskopatuhRow(sr model.ScanResult) map[string]string {
	row := map[string]string{}

	data := sr.ExtractedData
	if data == nil {
		return row
	}

	extracted, ok := data.(map[string]any)
	if !ok {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return row
		}
		var m map[string]any
		if err := json.Unmarshal(dataBytes, &m); err != nil {
			return row
		}
		extracted = m
	}

	// nama_paspor takes priority only when present; otherwise fall back to nama
	namaPaspor, _ := extracted["nama_paspor"].(string)
	nama, _ := extracted["nama"].(string)
	if namaPaspor != "" {
		row["NAMA LENGKAP"] = namaPaspor
	} else if nama != "" {
		row["NAMA LENGKAP"] = nama
	}
	if v, ok := extracted["no_paspor"].(string); ok {
		row["NO PASPOR"] = v
	}
	if v, ok := extracted["nik"].(string); ok {
		row["NO KTP / NIK"] = v
	}
	if v, ok := extracted["no_paspor"].(string); ok {
		row["NO PASPOR"] = v
	}
	if v, ok := extracted["nama_paspor"].(string); ok {
		row["NAMA LENGKAP"] = v
	}
	if v, ok := extracted["tempat_lahir"].(string); ok {
		row["TEMPAT LAHIR"] = v
	}
	if v, ok := extracted["tanggal_lahir"].(string); ok {
		row["TANGGAL LAHIR"] = v
	}
	if v, ok := extracted["jenis_kelamin"].(string); ok {
		row["JENIS KELAMIN"] = mapGender(v)
	}
	if v, ok := extracted["alamat"].(string); ok {
		row["ALAMAT"] = v
	}
	if v, ok := extracted["provinsi"].(string); ok {
		row["PROVINSI"] = v
	}
	if v, ok := extracted["kabupaten"].(string); ok {
		row["KABUPATEN / KOTA"] = v
	}
	if v, ok := extracted["kecamatan"].(string); ok {
		row["KECAMATAN"] = v
	}
	if v, ok := extracted["kelurahan"].(string); ok {
		row["KELURAHAN"] = v
	}
	if v, ok := extracted["no_telepon"].(string); ok {
		row["NO TELEPON"] = v
	}
	if v, ok := extracted["no_hp"].(string); ok {
		row["NO HP"] = v
	}
	if v, ok := extracted["kewarganegaraan"].(string); ok {
		row["KEWARGANEGARAAN"] = v
	}
	if v, ok := extracted["status_perkawinan"].(string); ok {
		row["STATUS PERKAWINAN"] = v
	}
	if v, ok := extracted["pendidikan"].(string); ok {
		row["PENDIDIKAN TERAKHIR"] = v
	}
	if v, ok := extracted["pekerjaan"].(string); ok {
		row["PEKERJAAN"] = v
	}
	if v, ok := extracted["golongan_darah"].(string); ok {
		row["GOLONGAN DARAH"] = v
	}
	if v, ok := extracted["tanggal_paspor"].(string); ok {
		row["TANGGAL TERBIT PASPOR"] = v
	}
	if v, ok := extracted["tanggal_expired"].(string); ok {
		row["TANGGAL EXPIRED PASPOR"] = v
	}
	if v, ok := extracted["kota_paspor"].(string); ok {
		row["TEMPAT TERBIT PASPOR"] = v
	}
	if v, ok := extracted["provider_visa"].(string); ok {
		row["PROVIDER VISA"] = v
	}
	if v, ok := extracted["no_visa"].(string); ok {
		row["NO VISA"] = v
	}
	if v, ok := extracted["tanggal_visa"].(string); ok {
		row["TANGGAL TERBIT VISA"] = v
	}
	if v, ok := extracted["tanggal_visa_akhir"].(string); ok {
		row["TANGGAL EXPIRED VISA"] = v
	}

	if normalized := sr.NormalizedData; normalized != nil {
		if n, ok := normalized.(map[string]any); ok {
			if v, ok := n["nama_ayah"].(string); ok && v != "" {
				row["NAMA AYAH"] = v
			}
		}
	}

	return row
}

func mapGender(g string) string {
	g = strings.ToLower(strings.TrimSpace(g))
	switch g {
	case "laki-laki", "laki", "laki2", "pria", "male", "lakilaki":
		return "Laki-Laki"
	case "perempuan", "wanita", "female", "cewek":
		return "Perempuan"
	default:
		return g
	}
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

// ExportRecordsExcel builds a Siskopatuh-format .xlsx from inline records (the
// scanner preview rows or group members the UI sends to /generate-excel),
// rather than re-querying scan results by package.
func (s *AIOCRService) ExportRecordsExcel(records []map[string]any) ([]byte, error) {
	return generateInlineSiskopatuhExcel(records)
}

func generateInlineSiskopatuhExcel(records []map[string]any) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Siskopatuh"
	_ = f.SetSheetName("Sheet1", sheetName)

	for i, colName := range siskopatuhColumns {
		cell := fmt.Sprintf("%s1", columnLetter(i+1))
		_ = f.SetCellValue(sheetName, cell, colName)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#2563EB"}},
	})
	_ = f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", columnLetter(len(siskopatuhColumns))), style)

	for i, rec := range records {
		row := i + 2
		data := siskopatuhRowFromRecord(rec)
		data["No"] = fmt.Sprintf("%d", i+1)
		for j, colName := range siskopatuhColumns {
			cell := fmt.Sprintf("%s%d", columnLetter(j+1), row)
			if val := data[colName]; val != "" {
				_ = f.SetCellValue(sheetName, cell, val)
			}
		}
	}

	for i := 1; i <= len(siskopatuhColumns); i++ {
		col := columnLetter(i)
		_ = f.SetColWidth(sheetName, col, col, 25)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}

// siskopatuhRowFromRecord maps a record to Siskopatuh columns. It accepts both
// the normalized OCR shape (no_identitas, gender, status_pernikahan, …) and the
// jamaah member shape (nik, jenis_kelamin, status_perkawinan, …) via aliases.
func siskopatuhRowFromRecord(rec map[string]any) map[string]string {
	row := map[string]string{}
	get := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := rec[k].(string); ok && strings.TrimSpace(v) != "" {
				return v
			}
		}
		return ""
	}

	if v := get("nama_paspor", "nama"); v != "" {
		row["NAMA LENGKAP"] = v
	}
	if v := get("gender", "jenis_kelamin"); v != "" {
		row["JENIS KELAMIN"] = mapGender(v)
	}
	if v := get("tempat_lahir"); v != "" {
		row["TEMPAT LAHIR"] = v
	}
	if v := get("tanggal_lahir"); v != "" {
		row["TANGGAL LAHIR"] = v
	}
	if v := get("alamat"); v != "" {
		row["ALAMAT"] = v
	}
	if v := get("provinsi"); v != "" {
		row["PROVINSI"] = v
	}
	if v := get("kabupaten"); v != "" {
		row["KABUPATEN / KOTA"] = v
	}
	if v := get("kecamatan"); v != "" {
		row["KECAMATAN"] = v
	}
	if v := get("kelurahan"); v != "" {
		row["KELURAHAN"] = v
	}
	if v := get("no_identitas", "nik"); v != "" {
		row["NO KTP / NIK"] = v
	}
	if v := get("no_paspor"); v != "" {
		row["NO PASPOR"] = v
	}
	if v := get("tanggal_paspor", "tanggal_terbit_paspor"); v != "" {
		row["TANGGAL TERBIT PASPOR"] = v
	}
	if v := get("tanggal_expired_paspor", "tanggal_expired"); v != "" {
		row["TANGGAL EXPIRED PASPOR"] = v
	}
	if v := get("kota_paspor"); v != "" {
		row["TEMPAT TERBIT PASPOR"] = v
	}
	if v := get("no_telepon"); v != "" {
		row["NO TELEPON"] = v
	}
	if v := get("no_hp"); v != "" {
		row["NO HP"] = v
	}
	if v := get("email"); v != "" {
		row["EMAIL"] = v
	}
	if v := get("kewarganegaraan"); v != "" {
		row["KEWARGANEGARAAN"] = v
	}
	if v := get("status_pernikahan", "status_perkawinan"); v != "" {
		row["STATUS PERKAWINAN"] = v
	}
	if v := get("pendidikan"); v != "" {
		row["PENDIDIKAN TERAKHIR"] = v
	}
	if v := get("pekerjaan"); v != "" {
		row["PEKERJAAN"] = v
	}
	if v := get("golongan_darah"); v != "" {
		row["GOLONGAN DARAH"] = v
	}
	if v := get("nama_ayah"); v != "" {
		row["NAMA AYAH"] = v
	}
	if v := get("no_visa"); v != "" {
		row["NO VISA"] = v
	}
	if v := get("tanggal_visa"); v != "" {
		row["TANGGAL TERBIT VISA"] = v
	}
	if v := get("tanggal_visa_akhir"); v != "" {
		row["TANGGAL EXPIRED VISA"] = v
	}
	if v := get("provider_visa"); v != "" {
		row["PROVIDER VISA"] = v
	}

	return row
}

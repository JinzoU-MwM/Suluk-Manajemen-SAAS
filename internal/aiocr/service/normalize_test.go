package service

import "testing"

func TestNormalizeNameMirrors(t *testing.T) {
	// Passport where the OCR put the name only in nama_paspor (the reported bug:
	// "Nama" column came out blank while "Nama Paspor" was filled).
	got := normalizeToSiskopatuh(
		ExtractedFields{NamaPaspor: "HENDRA MAHPUDIN", NoPaspor: "X8558076"}, "paspor")
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", got)
	}
	if m["nama"] != "HENDRA MAHPUDIN" || m["nama_paspor"] != "HENDRA MAHPUDIN" {
		t.Errorf("name not mirrored: nama=%v nama_paspor=%v", m["nama"], m["nama_paspor"])
	}

	// Normal case (Nama set) keeps both keys equal too.
	m2 := normalizeToSiskopatuh(
		ExtractedFields{Nama: "DWINTA DISTIANE", NoPaspor: "X8557911"}, "paspor").(map[string]any)
	if m2["nama"] != "DWINTA DISTIANE" || m2["nama_paspor"] != "DWINTA DISTIANE" {
		t.Errorf("name mismatch: nama=%v nama_paspor=%v", m2["nama"], m2["nama_paspor"])
	}

	// No name at all → neither key set (operator fills it).
	m3 := normalizeToSiskopatuh(ExtractedFields{NoPaspor: "X1"}, "paspor").(map[string]any)
	if _, ok := m3["nama"]; ok {
		t.Errorf("nama should be unset when no name extracted, got %v", m3["nama"])
	}
}

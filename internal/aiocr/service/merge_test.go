package service

import "testing"

func TestMergeIdentityRowsPassportVisa(t *testing.T) {
	// A passport and the matching visa (both carry the passport number) must
	// collapse into one jamaah row carrying both sets of fields.
	rows := []any{
		map[string]any{
			"nama": "AHMAD FAUZI", "no_paspor": "C1234567", "jenis_identitas": "PASPOR",
			"no_identitas": "C1234567", "tanggal_lahir": "1990-01-02",
		},
		map[string]any{
			"nama": "AHMAD FAUZI", "no_paspor": "C1234567", "jenis_identitas": "PASPOR",
			"no_visa": "V99", "provider_visa": "B2C", "tanggal_visa": "2026-05-01",
		},
	}
	out := mergeIdentityRows(rows)
	if len(out) != 1 {
		t.Fatalf("passport + visa should merge to 1 row, got %d", len(out))
	}
	m := out[0].(map[string]any)
	for k, want := range map[string]string{
		"no_paspor": "C1234567", "tanggal_lahir": "1990-01-02",
		"no_visa": "V99", "provider_visa": "B2C", "tanggal_visa": "2026-05-01",
	} {
		if got, _ := m[k].(string); got != want {
			t.Errorf("merged[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestMergeIdentityRowsKeepsDistinctPeople(t *testing.T) {
	rows := []any{
		map[string]any{"nama": "A", "no_paspor": "C1"},
		map[string]any{"nama": "B", "no_paspor": "C2"},
	}
	if out := mergeIdentityRows(rows); len(out) != 2 {
		t.Fatalf("two different passports must stay 2 rows, got %d", len(out))
	}
}

func TestMergeIdentityRowsKtpByName(t *testing.T) {
	// A KTP (no passport number) attaches to the passport row of the same name.
	rows := []any{
		map[string]any{"nama": "SITI AMINAH", "no_paspor": "X1", "jenis_identitas": "PASPOR"},
		map[string]any{"nama": "Siti Aminah", "no_identitas": "327301", "jenis_identitas": "NIK", "alamat": "JL MAWAR"},
	}
	out := mergeIdentityRows(rows)
	if len(out) != 1 {
		t.Fatalf("KTP should merge into the passport row by name, got %d rows", len(out))
	}
	if m := out[0].(map[string]any); m["alamat"] != "JL MAWAR" || m["no_paspor"] != "X1" {
		t.Errorf("merged row missing fields: %+v", m)
	}
}

func TestMergeIdentityRowsAmbiguousNameNotMerged(t *testing.T) {
	// Two different passports share a name; a passport-less row must NOT guess.
	rows := []any{
		map[string]any{"nama": "BUDI", "no_paspor": "X1"},
		map[string]any{"nama": "BUDI", "no_paspor": "X2"},
		map[string]any{"nama": "BUDI", "no_identitas": "327301", "jenis_identitas": "NIK"},
	}
	if out := mergeIdentityRows(rows); len(out) != 3 {
		t.Fatalf("ambiguous name must stay separate, got %d rows", len(out))
	}
}

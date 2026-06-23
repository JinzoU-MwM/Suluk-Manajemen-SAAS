package service

import "strings"

// Province / kabupaten normalization against the PropKab reference (see
// propkab_data.go). The template's dropdowns carry decorative spacing
// ("B A L I", "KAB. B A N T U L"); we match on a space/punctuation-stripped key
// and emit the canonical string verbatim so Siskopatuh accepts the cell.

// pkKey reduces a province/kabupaten name to a comparison key: uppercase, expand
// "KABUPATEN" -> "KAB" (KTPs spell it out; the template abbreviates), then drop
// every non-alphanumeric rune so spacing/punctuation differences disappear.
func pkKey(s string) string {
	t := strings.ToUpper(strings.TrimSpace(s))
	t = strings.ReplaceAll(t, "KABUPATEN", "KAB")
	var b strings.Builder
	for _, r := range t {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// provinceByKey maps a normalized key to the canonical province string, plus a
// few aliases for names the OCR spells differently than the template.
var provinceByKey = func() map[string]string {
	m := make(map[string]string, len(siskopatuhProvinces)+8)
	for _, p := range siskopatuhProvinces {
		m[pkKey(p)] = p
	}
	for alias, canon := range map[string]string{
		"YOGYAKARTA":               "D.I. YOGYAKARTA",
		"DIYOGYAKARTA":             "D.I. YOGYAKARTA",
		"DAERAHISTIMEWAYOGYAKARTA": "D.I. YOGYAKARTA",
		"JOGJAKARTA":               "D.I. YOGYAKARTA",
		"JAKARTA":                  "DKI JAKARTA",
		"DKI":                      "DKI JAKARTA",
		"BANGKABELITUNG":           "BANGKA BELITUNG",
		"KEPULAUANBANGKABELITUNG":  "BANGKA BELITUNG",
	} {
		m[alias] = canon
	}
	return m
}()

// kabupatenByProvince maps canonical province -> (normalized kabupaten key ->
// canonical kabupaten string).
var kabupatenByProvince = func() map[string]map[string]string {
	out := make(map[string]map[string]string, len(siskopatuhKabupaten))
	for prov, list := range siskopatuhKabupaten {
		mm := make(map[string]string, len(list))
		for _, k := range list {
			mm[pkKey(k)] = k
		}
		out[prov] = mm
	}
	return out
}()

// pkBareKey strips a leading administrative token (KAB / KOTA / their ADM
// variants) from an already-normalized key, so OCR text that omits the prefix
// ("BANTUL") can still match a canonical "KAB. B A N T U L".
func pkBareKey(key string) string {
	for _, p := range []string{"KOTAADMINISTRASI", "KOTAADM", "KABADM", "KOTA", "KAB"} {
		if strings.HasPrefix(key, p) {
			return strings.TrimPrefix(key, p)
		}
	}
	return key
}

// kabBareByProvince / kabBareGlobal index canonical kabupaten by their
// prefix-stripped bare key. A bare name is only trusted to resolve when it maps
// to exactly ONE canonical (e.g. "BANDUNG" is ambiguous between KAB. and KOTA
// BANDUNG, so it is left for the operator).
var (
	kabBareByProvince = map[string]map[string][]string{}
	kabBareGlobal     = map[string][]string{}
)

func init() {
	addUnique := func(m map[string][]string, key, canon string) {
		for _, c := range m[key] {
			if c == canon {
				return
			}
		}
		m[key] = append(m[key], canon)
	}
	for prov, list := range siskopatuhKabupaten {
		bp := map[string][]string{}
		for _, k := range list {
			bare := pkBareKey(pkKey(k))
			addUnique(bp, bare, k)
			addUnique(kabBareGlobal, bare, k)
		}
		kabBareByProvince[prov] = bp
	}
}

// mapProvinsi canonicalises a province name to its exact template value, or
// returns the cleaned uppercase input when no confident match exists.
func mapProvinsi(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	cleaned := strings.ReplaceAll(strings.ToUpper(s), "PROVINSI", "")
	cleaned = strings.ReplaceAll(cleaned, "PROPINSI", "")
	if c, ok := provinceByKey[pkKey(cleaned)]; ok {
		return c
	}
	return normUpper(s)
}

// mapKabupaten canonicalises a kabupaten/kota within a (canonical) province.
// If the province is unknown/blank it searches every province in template order
// (kabupaten names are nationally unique), then falls back to cleaned input.
func mapKabupaten(canonProvinsi, kab string) string {
	if strings.TrimSpace(kab) == "" {
		return ""
	}
	key := pkKey(kab)
	// 1. Exact match (with administrative prefix) inside the known province.
	if mm, ok := kabupatenByProvince[canonProvinsi]; ok {
		if c, ok := mm[key]; ok {
			return c
		}
	}
	// 2. Exact match across all provinces (province unknown/mismatched).
	for _, prov := range siskopatuhProvinces {
		if c, ok := kabupatenByProvince[prov][key]; ok {
			return c
		}
	}
	// 3. Prefix-less name, but only when it resolves to a single canonical.
	bare := pkBareKey(key)
	if cs := kabBareByProvince[canonProvinsi][bare]; len(cs) == 1 {
		return cs[0]
	}
	if cs := kabBareGlobal[bare]; len(cs) == 1 {
		return cs[0]
	}
	return normUpper(kab)
}

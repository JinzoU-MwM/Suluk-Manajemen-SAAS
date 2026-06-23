package service

import (
	"fmt"
	"strings"
)

// mergeIdentityRows combines the documents of a single jamaah into one row, so
// the scanner returns one person per row instead of one row per file. A passport
// and its visa both carry the passport number, so that is the primary key; a
// document without a passport number (a KTP, or a visa whose passport number was
// missed) is attached to a passport group by an exact, unambiguous name match,
// otherwise it stays on its own row.
//
// Fields merge gap-filling: the first row to claim a person is the base, and
// later rows only fill its empty/missing fields (so no extracted value is lost).
func mergeIdentityRows(data []any) []any {
	if len(data) <= 1 {
		return data
	}

	type grp struct {
		m   map[string]any
		raw any
	}
	var groups []*grp
	byPaspor := map[string]*grp{}
	var pending []*grp

	for _, item := range data {
		m, ok := item.(map[string]any)
		if !ok {
			groups = append(groups, &grp{raw: item})
			continue
		}
		if key := rowPasporKey(m); key != "" {
			if g := byPaspor[key]; g != nil {
				mergeRowInto(g.m, m)
			} else {
				g := &grp{m: m}
				byPaspor[key] = g
				groups = append(groups, g)
			}
			continue
		}
		pending = append(pending, &grp{m: m})
	}

	// Index the passport groups by name, but only names that are unique among
	// them, so a passport-less row never merges into the wrong person.
	nameCount := map[string]int{}
	nameGrp := map[string]*grp{}
	for _, g := range groups {
		if g.m == nil {
			continue
		}
		if n := rowNameKey(g.m); n != "" {
			nameCount[n]++
			nameGrp[n] = g
		}
	}
	for _, p := range pending {
		if n := rowNameKey(p.m); n != "" && nameCount[n] == 1 {
			mergeRowInto(nameGrp[n].m, p.m)
			continue
		}
		groups = append(groups, p)
	}

	out := make([]any, 0, len(groups))
	for _, g := range groups {
		if g.m != nil {
			out = append(out, g.m)
		} else {
			out = append(out, g.raw)
		}
	}
	return out
}

// rowNameKey is a row's normalized name (for cross-document person matching).
func rowNameKey(m map[string]any) string {
	return normUpper(fieldGetter{m}.first("nama", "nama_paspor"))
}

// mergeRowInto copies every non-empty src field into dst, without overwriting a
// value dst already has (first document wins on conflict).
func mergeRowInto(dst, src map[string]any) {
	for k, v := range src {
		if isEmptyVal(v) {
			continue
		}
		if ev, ok := dst[k]; !ok || isEmptyVal(ev) {
			dst[k] = v
		}
	}
}

func isEmptyVal(v any) bool {
	return v == nil || strings.TrimSpace(fmt.Sprintf("%v", v)) == ""
}

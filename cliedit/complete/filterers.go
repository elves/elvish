package complete

import "strings"

func FilterPrefix(ctxName, seed string, items []RawItem) []RawItem {
	var filtered []RawItem
	for _, cand := range items {
		if strings.HasPrefix(cand.String(), seed) {
			filtered = append(filtered, cand)
		}
	}
	return filtered
}

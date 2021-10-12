package complete

// FilterPrefix filters raw items by prefix. It can be used as a Filterer in
// Config.
func FilterPrefix(ctxName, seed string, items []RawItem) []RawItem {
	var filtered []RawItem
	for _, cand := range items {
		if cand.HasPrefix(seed) {
			filtered = append(filtered, cand)
		}
	}
	return filtered
}

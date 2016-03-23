package edit

import "sort"

// styled is a piece of text with style.
type styled struct {
	text  string
	style string
}

type styleds []styled

func (s styleds) Len() int           { return len(s) }
func (s styleds) Less(i, j int) bool { return s[i].text < s[j].text }
func (s styleds) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func sortStyleds(s []styled) {
	sort.Sort(styleds(s))
}

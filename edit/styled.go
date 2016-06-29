package edit

import "sort"

// styled is a piece of text with style.
type styled struct {
	text  string
	style string
}

func unstyled(s string) styled {
	return styled{s, ""}
}

func (s *styled) addStyle(st string) {
	s.style = joinStyle(s.style, st)
}

type styleds []styled

func (s styleds) Len() int           { return len(s) }
func (s styleds) Less(i, j int) bool { return s[i].text < s[j].text }
func (s styleds) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func sortStyleds(s []styled) {
	sort.Sort(styleds(s))
}

package styled

import (
	"strings"
)

// Transform transforms a Text according to a transformer. It does nothing if
// the transformer is not valid.
func Transform(t Text, transformer string) Text {
	f := FindTransformer(transformer)
	if f == nil {
		return t
	}
	t = t.Clone()
	for _, seg := range t {
		f(seg)
	}
	return t
}

// FindTransformer finds the named transformer, a function that mutates a
// *Segment. If the name is not a valid transformer, it returns nil.
func FindTransformer(name string) func(*Segment) {
	switch {
	// Catch special colors early
	case name == "default":
		return func(s *Segment) { s.Foreground = "" }
	case name == "bg-default":
		return func(s *Segment) { s.Background = "" }
	case strings.HasPrefix(name, "bg-"):
		if color := name[len("bg-"):]; isValidColorName(color) {
			return func(s *Segment) { s.Background = color }
		}
	case strings.HasPrefix(name, "no-"):
		if f := boolFieldAccessor(name[len("no-"):]); f != nil {
			return func(s *Segment) { *f(s) = false }
		}
	case strings.HasPrefix(name, "toggle-"):
		if f := boolFieldAccessor(name[len("toggle-"):]); f != nil {
			return func(s *Segment) {
				p := f(s)
				*p = !*p
			}
		}
	default:
		if isValidColorName(name) {
			return func(s *Segment) { s.Foreground = name }
		}
		if f := boolFieldAccessor(name); f != nil {
			return func(s *Segment) { *f(s) = true }
		}
	}
	return nil
}

func boolFieldAccessor(name string) func(*Segment) *bool {
	switch name {
	case "bold":
		return func(s *Segment) *bool { return &s.Bold }
	case "dim":
		return func(s *Segment) *bool { return &s.Dim }
	case "italic":
		return func(s *Segment) *bool { return &s.Italic }
	case "underlined":
		return func(s *Segment) *bool { return &s.Underlined }
	case "blink":
		return func(s *Segment) *bool { return &s.Blink }
	case "inverse":
		return func(s *Segment) *bool { return &s.Inverse }
	default:
		return nil
	}
}

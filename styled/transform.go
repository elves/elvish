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
		f(&seg.Style)
	}
	return t
}

// FindTransformer finds the named transformer, a function that mutates a
// *Style. If the name is not a valid transformer, it returns nil.
func FindTransformer(name string) func(*Style) {
	switch {
	// Catch special colors early
	case name == "default":
		return func(s *Style) { s.Foreground = "" }
	case name == "bg-default":
		return func(s *Style) { s.Background = "" }
	case strings.HasPrefix(name, "bg-"):
		if color := name[len("bg-"):]; isValidColorName(color) {
			return func(s *Style) { s.Background = color }
		}
	case strings.HasPrefix(name, "no-"):
		if f := boolFieldAccessor(name[len("no-"):]); f != nil {
			return func(s *Style) { *f(s) = false }
		}
	case strings.HasPrefix(name, "toggle-"):
		if f := boolFieldAccessor(name[len("toggle-"):]); f != nil {
			return func(s *Style) {
				p := f(s)
				*p = !*p
			}
		}
	default:
		if isValidColorName(name) {
			return func(s *Style) { s.Foreground = name }
		}
		if f := boolFieldAccessor(name); f != nil {
			return func(s *Style) { *f(s) = true }
		}
	}
	return nil
}

func boolFieldAccessor(name string) func(*Style) *bool {
	switch name {
	case "bold":
		return func(s *Style) *bool { return &s.Bold }
	case "dim":
		return func(s *Style) *bool { return &s.Dim }
	case "italic":
		return func(s *Style) *bool { return &s.Italic }
	case "underlined":
		return func(s *Style) *bool { return &s.Underlined }
	case "blink":
		return func(s *Style) *bool { return &s.Blink }
	case "inverse":
		return func(s *Style) *bool { return &s.Inverse }
	default:
		return nil
	}
}

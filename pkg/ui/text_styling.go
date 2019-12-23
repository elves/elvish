package ui

import (
	"strings"
)

// Styling specifies how to change a Style. It can also be applied to a Segment
// or Text.
type Styling interface{ transform(*Style) }

// StyleText returns a new Text with the given Styling's applied. It does not
// modify the given Text.
func StyleText(t Text, ts ...Styling) Text {
	newt := make(Text, len(t))
	for i, seg := range t {
		newt[i] = StyleSegment(seg, ts...)
	}
	return newt
}

// StyleSegment returns a new Segment with the given Styling's applied. It does
// not modify the given Segment.
func StyleSegment(seg *Segment, ts ...Styling) *Segment {
	return &Segment{Text: seg.Text, Style: ApplyStyling(seg.Style, ts...)}
}

// ApplyStyling returns a new Style with the given Styling's applied.
func ApplyStyling(s Style, ts ...Styling) Style {
	for _, t := range ts {
		if t != nil {
			t.transform(&s)
		}
	}
	return s
}

// Stylings joins several transformers into one.
func Stylings(ts ...Styling) Styling { return jointStyling(ts) }

var (
	FgDefault = setForeground{nil}

	FgBlack   = setForeground{Black}
	FgRed     = setForeground{Red}
	FgGreen   = setForeground{Green}
	FgYellow  = setForeground{Yellow}
	FgBlue    = setForeground{Blue}
	FgMagenta = setForeground{Magenta}
	FgCyan    = setForeground{Cyan}
	FgWhite   = setForeground{White}

	FgBrightBlack   = setForeground{BrightBlack}
	FgBrightRed     = setForeground{BrightRed}
	FgBrightGreen   = setForeground{BrightGreen}
	FgBrightYellow  = setForeground{BrightYellow}
	FgBrightBlue    = setForeground{BrightBlue}
	FgBrightMagenta = setForeground{BrightMagenta}
	FgBrightCyan    = setForeground{BrightCyan}
	FgBrightWhite   = setForeground{BrightWhite}

	BgDefault = setBackground{nil}

	BgBlack   = setBackground{Black}
	BgRed     = setBackground{Red}
	BgGreen   = setBackground{Green}
	BgYellow  = setBackground{Yellow}
	BgBlue    = setBackground{Blue}
	BgMagenta = setBackground{Magenta}
	BgCyan    = setBackground{Cyan}
	BgWhite   = setBackground{White}

	BgBrightBlack   = setBackground{BrightBlack}
	BgBrightRed     = setBackground{BrightRed}
	BgBrightGreen   = setBackground{BrightGreen}
	BgBrightYellow  = setBackground{BrightYellow}
	BgBrightBlue    = setBackground{BrightBlue}
	BgBrightMagenta = setBackground{BrightMagenta}
	BgBrightCyan    = setBackground{BrightCyan}
	BgBrightWhite   = setBackground{BrightWhite}

	Bold       = boolOn(accessBold)
	Dim        = boolOn(accessDim)
	Italic     = boolOn(accessItalic)
	Underlined = boolOn(accessUnderlined)
	Blink      = boolOn(accessBlink)
	Inverse    = boolOn(accessInverse)

	NoBold       = boolOff(accessBold)
	NoDim        = boolOff(accessDim)
	NoItalic     = boolOff(accessItalic)
	NoUnderlined = boolOff(accessUnderlined)
	NoBlink      = boolOff(accessBlink)
	NoInverse    = boolOff(accessInverse)

	ToggleBold       = boolToggle(accessBold)
	ToggleDim        = boolToggle(accessDim)
	ToggleItalic     = boolToggle(accessItalic)
	ToggleUnderlined = boolToggle(accessUnderlined)
	ToggleBlink      = boolToggle(accessBlink)
	ToggleInverse    = boolToggle(accessInverse)
)

type setForeground struct{ c Color }
type setBackground struct{ c Color }
type boolOn func(*Style) *bool
type boolOff func(*Style) *bool
type boolToggle func(*Style) *bool

func (t setForeground) transform(s *Style) { s.Foreground = t.c }
func (t setBackground) transform(s *Style) { s.Background = t.c }
func (t boolOn) transform(s *Style)        { *t(s) = true }
func (t boolOff) transform(s *Style)       { *t(s) = false }
func (t boolToggle) transform(s *Style)    { p := t(s); *p = !*p }

func accessBold(s *Style) *bool       { return &s.Bold }
func accessDim(s *Style) *bool        { return &s.Dim }
func accessItalic(s *Style) *bool     { return &s.Italic }
func accessUnderlined(s *Style) *bool { return &s.Underlined }
func accessBlink(s *Style) *bool      { return &s.Blink }
func accessInverse(s *Style) *bool    { return &s.Inverse }

type jointStyling []Styling

func (t jointStyling) transform(s *Style) {
	for _, t := range t {
		t.transform(s)
	}
}

// ParseStyling parses a text representation of Styling, which are kebab
// case counterparts to the names of the builtin Styling's. For example,
// ToggleInverse is expressed as "toggle-inverse".
//
// Multiple stylings can be joined by spaces, which is equivalent to calling
// Stylings.
//
// If the given string is invalid, ParseStyling returns nil.
func ParseStyling(s string) Styling {
	if !strings.ContainsRune(s, ' ') {
		return parseOneStyling(s)
	}
	var joint jointStyling
	for _, subs := range strings.Split(s, " ") {
		parsed := parseOneStyling(subs)
		if parsed == nil {
			return nil
		}
		joint = append(joint, parseOneStyling(subs))
	}
	return joint
}

var boolFieldAccessor = map[string]func(*Style) *bool{
	"bold":       accessBold,
	"dim":        accessDim,
	"italic":     accessItalic,
	"underlined": accessUnderlined,
	"blink":      accessBlink,
	"inverse":    accessInverse,
}

func parseOneStyling(name string) Styling {
	switch {
	case name == "default" || name == "fg-default":
		return FgDefault
	case strings.HasPrefix(name, "fg-"):
		if color := parseColor(name[len("fg-"):]); color != nil {
			return setForeground{color}
		}
	case name == "bg-default":
		return BgDefault
	case strings.HasPrefix(name, "bg-"):
		if color := parseColor(name[len("bg-"):]); color != nil {
			return setBackground{color}
		}
	case strings.HasPrefix(name, "no-"):
		if f, ok := boolFieldAccessor[name[len("no-"):]]; ok {
			return boolOff(f)
		}
	case strings.HasPrefix(name, "toggle-"):
		if f, ok := boolFieldAccessor[name[len("toggle-"):]]; ok {
			return boolToggle(f)
		}
	default:
		if f, ok := boolFieldAccessor[name]; ok {
			return boolOn(f)
		}
		if color := parseColor(name); color != nil {
			return setForeground{color}
		}
	}
	return nil
}

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

// Common stylings.
var (
	FgDefault Styling = setForeground{nil}

	FgBlack   Styling = setForeground{Black}
	FgRed     Styling = setForeground{Red}
	FgGreen   Styling = setForeground{Green}
	FgYellow  Styling = setForeground{Yellow}
	FgBlue    Styling = setForeground{Blue}
	FgMagenta Styling = setForeground{Magenta}
	FgCyan    Styling = setForeground{Cyan}
	FgWhite   Styling = setForeground{White}

	FgBrightBlack   Styling = setForeground{BrightBlack}
	FgBrightRed     Styling = setForeground{BrightRed}
	FgBrightGreen   Styling = setForeground{BrightGreen}
	FgBrightYellow  Styling = setForeground{BrightYellow}
	FgBrightBlue    Styling = setForeground{BrightBlue}
	FgBrightMagenta Styling = setForeground{BrightMagenta}
	FgBrightCyan    Styling = setForeground{BrightCyan}
	FgBrightWhite   Styling = setForeground{BrightWhite}

	BgDefault Styling = setBackground{nil}

	BgBlack   Styling = setBackground{Black}
	BgRed     Styling = setBackground{Red}
	BgGreen   Styling = setBackground{Green}
	BgYellow  Styling = setBackground{Yellow}
	BgBlue    Styling = setBackground{Blue}
	BgMagenta Styling = setBackground{Magenta}
	BgCyan    Styling = setBackground{Cyan}
	BgWhite   Styling = setBackground{White}

	BgBrightBlack   Styling = setBackground{BrightBlack}
	BgBrightRed     Styling = setBackground{BrightRed}
	BgBrightGreen   Styling = setBackground{BrightGreen}
	BgBrightYellow  Styling = setBackground{BrightYellow}
	BgBrightBlue    Styling = setBackground{BrightBlue}
	BgBrightMagenta Styling = setBackground{BrightMagenta}
	BgBrightCyan    Styling = setBackground{BrightCyan}
	BgBrightWhite   Styling = setBackground{BrightWhite}

	Bold       Styling = boolOn(accessBold)
	Dim        Styling = boolOn(accessDim)
	Italic     Styling = boolOn(accessItalic)
	Underlined Styling = boolOn(accessUnderlined)
	Blink      Styling = boolOn(accessBlink)
	Inverse    Styling = boolOn(accessInverse)

	NoBold       Styling = boolOff(accessBold)
	NoDim        Styling = boolOff(accessDim)
	NoItalic     Styling = boolOff(accessItalic)
	NoUnderlined Styling = boolOff(accessUnderlined)
	NoBlink      Styling = boolOff(accessBlink)
	NoInverse    Styling = boolOff(accessInverse)

	ToggleBold       Styling = boolToggle(accessBold)
	ToggleDim        Styling = boolToggle(accessDim)
	ToggleItalic     Styling = boolToggle(accessItalic)
	ToggleUnderlined Styling = boolToggle(accessUnderlined)
	ToggleBlink      Styling = boolToggle(accessBlink)
	ToggleInverse    Styling = boolToggle(accessInverse)
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

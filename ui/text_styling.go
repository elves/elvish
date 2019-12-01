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

// JoinStylings joins several transformers into one.
func JoinStylings(ts ...Styling) Styling { return jointStyling(ts) }

var (
	Black        Styling = setForeground("black")
	Red                  = setForeground("red")
	Green                = setForeground("green")
	Yellow               = setForeground("yellow")
	Blue                 = setForeground("blue")
	Magenta              = setForeground("magenta")
	Cyan                 = setForeground("cyan")
	LightGray            = setForeground("lightgray")
	Gray                 = setForeground("gray")
	LightRed             = setForeground("lightred")
	LightGreen           = setForeground("lightgreen")
	LightYellow          = setForeground("lightyellow")
	LightBlue            = setForeground("lightblue")
	LightMagenta         = setForeground("lightmagenta")
	LightCyan            = setForeground("lightcyan")
	White                = setForeground("white")

	FgDefault      = setForeground("")
	FgBlack        = setForeground("black")
	FgRed          = setForeground("red")
	FgGreen        = setForeground("green")
	FgYellow       = setForeground("yellow")
	FgBlue         = setForeground("blue")
	FgMagenta      = setForeground("magenta")
	FgCyan         = setForeground("cyan")
	FgLightGray    = setForeground("lightgray")
	FgGray         = setForeground("gray")
	FgLightRed     = setForeground("lightred")
	FgLightGreen   = setForeground("lightgreen")
	FgLightYellow  = setForeground("lightyellow")
	FgLightBlue    = setForeground("lightblue")
	FgLightMagenta = setForeground("lightmagenta")
	FgLightCyan    = setForeground("lightcyan")
	FgWhite        = setForeground("white")

	BgDefault      = setBackground("default")
	BgBlack        = setBackground("black")
	BgRed          = setBackground("red")
	BgGreen        = setBackground("green")
	BgYellow       = setBackground("yellow")
	BgBlue         = setBackground("blue")
	BgMagenta      = setBackground("magenta")
	BgCyan         = setBackground("cyan")
	BgLightGray    = setBackground("lightgray")
	BgGray         = setBackground("gray")
	BgLightRed     = setBackground("lightred")
	BgLightGreen   = setBackground("lightgreen")
	BgLightYellow  = setBackground("lightyellow")
	BgLightBlue    = setBackground("lightblue")
	BgLightMagenta = setBackground("lightmagenta")
	BgLightCyan    = setBackground("lightcyan")
	BgWhite        = setBackground("white")

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

type setForeground string
type setBackground string
type boolOn func(*Style) *bool
type boolOff func(*Style) *bool
type boolToggle func(*Style) *bool

func (t setForeground) transform(s *Style) { s.Foreground = string(t) }
func (t setBackground) transform(s *Style) { s.Background = string(t) }
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
// JoinStylings.
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
		if color := name[len("fg-"):]; isValidColorName(color) {
			return setForeground(color)
		}
	case isValidColorName(name):
		return setForeground(name)
	case name == "bg-default":
		return BgDefault
	case strings.HasPrefix(name, "bg-"):
		if color := name[len("bg-"):]; isValidColorName(color) {
			return setBackground(color)
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
	}
	return nil
}

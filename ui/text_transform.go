package ui

import (
	"strings"
)

// Transformer specifies how to transform a Style, Segment or Text.
type Transformer interface{ transform(*Style) }

// Transform applies the given Transformer's to a Text. It does not mutate the
// given Text.
func Transform(t Text, ts ...Transformer) Text {
	newt := make(Text, len(t))
	for i, seg := range t {
		newt[i] = TransformSegment(seg, ts...)
	}
	return newt
}

// TransformSegment transforms a Segment according to the given Transformer's.
// It does not mutate the given Segment.
func TransformSegment(seg *Segment, ts ...Transformer) *Segment {
	return &Segment{Text: seg.Text, Style: TransformStyle(seg.Style, ts...)}
}

// TransformStyle transforms a Style according to the given Transformer's.
func TransformStyle(s Style, ts ...Transformer) Style {
	for _, t := range ts {
		if t != nil {
			t.transform(&s)
		}
	}
	return s
}

// JoinTransformers joins several transformers into one.
func JoinTransformers(ts ...Transformer) Transformer { return jointTransformer(ts) }

var (
	Black        Transformer = setForeground("black")
	Red                      = setForeground("red")
	Green                    = setForeground("green")
	Yellow                   = setForeground("yellow")
	Blue                     = setForeground("blue")
	Magenta                  = setForeground("magenta")
	Cyan                     = setForeground("cyan")
	LightGray                = setForeground("lightgray")
	Gray                     = setForeground("gray")
	LightRed                 = setForeground("lightred")
	LightGreen               = setForeground("lightgreen")
	LightYellow              = setForeground("lightyellow")
	LightBlue                = setForeground("lightblue")
	LightMagenta             = setForeground("lightmagenta")
	LightCyan                = setForeground("lightcyan")
	White                    = setForeground("white")

	BlackForeground        = setForeground("black")
	RedForeground          = setForeground("red")
	GreenForeground        = setForeground("green")
	YellowForeground       = setForeground("yellow")
	BlueForeground         = setForeground("blue")
	MagentaForeground      = setForeground("magenta")
	CyanForeground         = setForeground("cyan")
	LightGrayForeground    = setForeground("lightgray")
	GrayForeground         = setForeground("gray")
	LightRedForeground     = setForeground("lightred")
	LightGreenForeground   = setForeground("lightgreen")
	LightYellowForeground  = setForeground("lightyellow")
	LightBlueForeground    = setForeground("lightblue")
	LightMagentaForeground = setForeground("lightmagenta")
	LightCyanForeground    = setForeground("lightcyan")
	WhiteForeground        = setForeground("white")

	BlackBackground        = setBackground("black")
	RedBackground          = setBackground("red")
	GreenBackground        = setBackground("green")
	YellowBackground       = setBackground("yellow")
	BlueBackground         = setBackground("blue")
	MagentaBackground      = setBackground("magenta")
	CyanBackground         = setBackground("cyan")
	LightGrayBackground    = setBackground("lightgray")
	GrayBackground         = setBackground("gray")
	LightRedBackground     = setBackground("lightred")
	LightGreenBackground   = setBackground("lightgreen")
	LightYellowBackground  = setBackground("lightyellow")
	LightBlueBackground    = setBackground("lightblue")
	LightMagentaBackground = setBackground("lightmagenta")
	LightCyanBackground    = setBackground("lightcyan")
	WhiteBackground        = setBackground("white")

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

type jointTransformer []Transformer

func (t jointTransformer) transform(s *Style) {
	for _, t := range t {
		t.transform(s)
	}
}

// TransformText transforms a Text according to a transformer. It does nothing
// if the transformer is not valid.
func TransformText(t Text, transformer string) Text {
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

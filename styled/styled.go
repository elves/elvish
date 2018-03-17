package styled

import "github.com/elves/elvish/util"

type TextStyle struct {
	bold       *bool
	dim        *bool
	italic     *bool
	underlined *bool
	blink      *bool
	inverse    *bool
}

type Style struct {
	Foreground *Color
	Background *Color
	TextStyle
}

func TextStyleFromMap(m map[string]interface{}) TextStyle {
	b := func(key string) *bool {
		if b, ok := m[key]; ok {
			if b, ok := b.(bool); ok {
				return &b
			}
		}
		return nil
	}

	return TextStyle{
		bold:       b("bold"),
		dim:        b("dim"),
		italic:     b("italic"),
		underlined: b("underlined"),
		blink:      b("blink"),
		inverse:    b("inverse"),
	}
}

func (s TextStyle) Merge(o TextStyle) TextStyle {
	if o.bold != nil {
		s.bold = o.bold
	}
	if o.dim != nil {
		s.dim = o.dim
	}
	if o.italic != nil {
		s.italic = o.italic
	}
	if o.underlined != nil {
		s.underlined = o.underlined
	}
	if o.blink != nil {
		s.blink = o.blink
	}
	if o.inverse != nil {
		s.inverse = o.inverse
	}

	return s
}

func ForegroundColorFromOptions(options map[string]interface{}) *Color {
	return colorFromOptions(options, "fg-color")
}
func BackgroundColorFromOptions(options map[string]interface{}) *Color {
	return colorFromOptions(options, "bg-color")
}

func (s Style) ToMap() map[string]interface{} {
	c := func(c *Color) Color {
		if c != nil {
			return *c
		}
		return ColorDefault
	}
	b := func(b *bool) bool { return b != nil && *b }

	return map[string]interface{}{
		"fg-color":   c(s.Foreground),
		"bg-color":   c(s.Background),
		"bold":       b(s.bold),
		"dim":        b(s.dim),
		"italic":     b(s.italic),
		"underlined": b(s.underlined),
		"blink":      b(s.blink),
		"inverse":    b(s.inverse),
	}
}

func colorFromOptions(options map[string]interface{}, which string) *Color {
	if col, ok := options[which]; ok {
		if colString, ok := col.(string); ok {
			col, err := GetColorFromString(colString)
			if err != nil {
				util.Throw(err)
			}

			return &col
		}
	}

	return nil
}

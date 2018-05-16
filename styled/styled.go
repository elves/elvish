package styled

import (
	"fmt"

	"github.com/elves/elvish/eval/vals"
)

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

func TextStyleFromMap(m map[string]interface{}) (*TextStyle, error) {
	b := func(key string) (*bool, error) {
		if b, ok := m[key]; ok {
			if bl, ok := b.(bool); ok {
				return &bl, nil
			} else {
				return nil, fmt.Errorf("'%s' must be a boolean value; got %s", key, vals.Kind(b))
			}
		}
		return nil, nil
	}

	bold, err := b("bold")
	if err != nil {
		return nil, err
	}

	dim, err := b("dim")
	if err != nil {
		return nil, err
	}

	italic, err := b("italic")
	if err != nil {
		return nil, err
	}

	underlined, err := b("underlined")
	if err != nil {
		return nil, err
	}

	blink, err := b("blink")
	if err != nil {
		return nil, err
	}

	inverse, err := b("inverse")
	if err != nil {
		return nil, err
	}

	return &TextStyle{
		bold:       bold,
		dim:        dim,
		italic:     italic,
		underlined: underlined,
		blink:      blink,
		inverse:    inverse,
	}, nil
}

func (s TextStyle) Merge(o *TextStyle) TextStyle {
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

func ForegroundColorFromOptions(options map[string]interface{}) (*Color, error) {
	return colorFromOptions(options, "fg-color")
}
func BackgroundColorFromOptions(options map[string]interface{}) (*Color, error) {
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

func colorFromOptions(options map[string]interface{}, which string) (*Color, error) {
	if col, ok := options[which]; ok {
		if colString, ok := col.(string); ok {
			col, err := GetColorFromString(colString)
			if err != nil {
				return nil, err
			}

			return &col, nil
		}
	}

	return nil, nil
}

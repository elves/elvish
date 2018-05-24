package styled

import (
	"fmt"
)

type Style struct {
	Foreground string
	Background string
	Bold       bool
	Dim        bool
	Italic     bool
	Underlined bool
	Blink      bool
	Inverse    bool
}

func EmptyStyle() Style {
	return Style{
		Foreground: "default",
		Background: "default",
	}
}

func (s *Style) ImportFromOptions(options map[string]interface{}) error {
	assignColor := func(key string, assign func(string)) error {
		if c, ok := options[key]; ok {
			if c, ok := c.(string); ok && IsValidColorString(c) {
				assign(c)
			} else {
				return fmt.Errorf("value to option '%s' must be a valid color string", key)
			}
		}
		return nil
	}
	assignBool := func(key string, assign func(bool)) error {
		if b, ok := options[key]; ok {
			if b, ok := b.(bool); ok {
				assign(b)
			} else {
				return fmt.Errorf("value to option '%s' must be a bool value", key)
			}
		}

		return nil
	}

	if err := assignColor("fg-color", func(c string) { s.Foreground = c }); err != nil {
		return err
	}
	if err := assignColor("bg-color", func(c string) { s.Background = c }); err != nil {
		return err
	}
	if err := assignBool("bold", func(b bool) { s.Bold = b }); err != nil {
		return err
	}
	if err := assignBool("dim", func(b bool) { s.Dim = b }); err != nil {
		return err
	}
	if err := assignBool("italic", func(b bool) { s.Italic = b }); err != nil {
		return err
	}
	if err := assignBool("underlined", func(b bool) { s.Underlined = b }); err != nil {
		return err
	}
	if err := assignBool("blink", func(b bool) { s.Blink = b }); err != nil {
		return err
	}
	if err := assignBool("inverse", func(b bool) { s.Inverse = b }); err != nil {
		return err
	}

	return nil
}

func (s Style) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"fg-color":   s.Foreground,
		"bg-color":   s.Background,
		"bold":       s.Bold,
		"dim":        s.Dim,
		"italic":     s.Italic,
		"underlined": s.Underlined,
		"blink":      s.Blink,
		"inverse":    s.Inverse,
	}
}

func IsValidColorString(col string) bool {
	switch col {
	case
		"default",
		"black",
		"red",
		"green",
		"yellow",
		"blue",
		"magenta",
		"cyan",
		"lightgray",
		"gray",
		"lightred",
		"lightgreen",
		"lightyellow",
		"lightblue",
		"lightmagenta",
		"lightcyan",
		"white":
		return true

	default:
		return false
	}
}

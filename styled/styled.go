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

func (s *Style) ImportFromOptions(options map[string]interface{}) error {
	assignColor := func(key string, colorField *string) error {
		if c, ok := options[key]; ok {
			if c, ok := c.(string); ok && isValidColorName(c) {
				if c == "default" {
					*colorField = ""
				} else {
					*colorField = c
				}
			} else {
				return fmt.Errorf("value to option '%s' must be a valid color string", key)
			}
		}
		return nil
	}
	assignBool := func(key string, attrField *bool) error {
		if b, ok := options[key]; ok {
			if b, ok := b.(bool); ok {
				*attrField = b
			} else {
				return fmt.Errorf("value to option '%s' must be a bool value", key)
			}
		}

		return nil
	}

	if err := assignColor("fg-color", &s.Foreground); err != nil {
		return err
	}
	if err := assignColor("bg-color", &s.Background); err != nil {
		return err
	}
	if err := assignBool("bold", &s.Bold); err != nil {
		return err
	}
	if err := assignBool("dim", &s.Dim); err != nil {
		return err
	}
	if err := assignBool("italic", &s.Italic); err != nil {
		return err
	}
	if err := assignBool("underlined", &s.Underlined); err != nil {
		return err
	}
	if err := assignBool("blink", &s.Blink); err != nil {
		return err
	}
	if err := assignBool("inverse", &s.Inverse); err != nil {
		return err
	}

	return nil
}

func isValidColorName(col string) bool {
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

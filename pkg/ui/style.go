package ui

import (
	"fmt"
	"strconv"
	"strings"
)

// Style specifies how something (mostly a string) shall be displayed.
type Style struct {
	Foreground Color
	Background Color
	Bold       bool
	Dim        bool
	Italic     bool
	Underlined bool
	Blink      bool
	Inverse    bool
}

// Color represents a color.
type Color interface {
	fgSGR() string
	bgSGR() string
	String() string
}

// Builtin ANSI colors.
var (
	Black   Color = ansiColor(0)
	Red     Color = ansiColor(1)
	Green   Color = ansiColor(2)
	Yellow  Color = ansiColor(3)
	Blue    Color = ansiColor(4)
	Magenta Color = ansiColor(5)
	Cyan    Color = ansiColor(6)
	White   Color = ansiColor(7)

	BrightBlack   Color = ansiBrightColor(0)
	BrightRed     Color = ansiBrightColor(1)
	BrightGreen   Color = ansiBrightColor(2)
	BrightYellow  Color = ansiBrightColor(3)
	BrightBlue    Color = ansiBrightColor(4)
	BrightMagenta Color = ansiBrightColor(5)
	BrightCyan    Color = ansiBrightColor(6)
	BrightWhite   Color = ansiBrightColor(7)
)

// XTerm256Color returns a color from the xterm 256-color palette.
func XTerm256Color(i uint8) Color { return xterm256Color(i) }

// TrueColor returns a 24-bit true color.
func TrueColor(r, g, b uint8) Color { return trueColor{r, g, b} }

var colorNames = []string{
	"black", "red", "green", "yellow",
	"blue", "magenta", "cyan", "white",
}

var colorByName = map[string]Color{
	"black":   Black,
	"red":     Red,
	"green":   Green,
	"yellow":  Yellow,
	"blue":    Blue,
	"magenta": Magenta,
	"cyan":    Cyan,
	"white":   White,

	"bright-black":   BrightBlack,
	"bright-red":     BrightRed,
	"bright-green":   BrightGreen,
	"bright-yellow":  BrightYellow,
	"bright-blue":    BrightBlue,
	"bright-magenta": BrightMagenta,
	"bright-cyan":    BrightCyan,
	"bright-white":   BrightWhite,
}

type ansiColor uint8

func (c ansiColor) fgSGR() string  { return strconv.Itoa(30 + int(c)) }
func (c ansiColor) bgSGR() string  { return strconv.Itoa(40 + int(c)) }
func (c ansiColor) String() string { return colorNames[c] }

type ansiBrightColor uint8

func (c ansiBrightColor) fgSGR() string  { return strconv.Itoa(90 + int(c)) }
func (c ansiBrightColor) bgSGR() string  { return strconv.Itoa(100 + int(c)) }
func (c ansiBrightColor) String() string { return "bright-" + colorNames[c] }

type xterm256Color uint8

func (c xterm256Color) fgSGR() string  { return "38;5;" + strconv.Itoa(int(c)) }
func (c xterm256Color) bgSGR() string  { return "48;5;" + strconv.Itoa(int(c)) }
func (c xterm256Color) String() string { return "color" + strconv.Itoa(int(c)) }

type trueColor struct{ r, g, b uint8 }

func (c trueColor) fgSGR() string { return "38;2;" + c.rgbSGR() }
func (c trueColor) bgSGR() string { return "48;2;" + c.rgbSGR() }

func (c trueColor) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b)
}

func (c trueColor) rgbSGR() string {
	return fmt.Sprintf("%d;%d;%d", c.r, c.g, c.b)
}

// ImportFromOptions assigns all recognized values from a map to the current
// Style.
func (s *Style) ImportFromOptions(options map[string]interface{}) error {
	assignColor := func(val interface{}, colorField *Color) string {
		if val == "default" {
			*colorField = nil
			return ""
		} else if s, ok := val.(string); ok {
			color := parseColor(s)
			if color != nil {
				*colorField = color
				return ""
			}
		}
		return "valid color string"
	}
	assignBool := func(val interface{}, attrField *bool) string {
		if b, ok := val.(bool); ok {
			*attrField = b
		} else {
			return "bool value"
		}
		return ""
	}

	for k, v := range options {
		var need string

		switch k {
		case "fg-color":
			need = assignColor(v, &s.Foreground)
		case "bg-color":
			need = assignColor(v, &s.Background)
		case "bold":
			need = assignBool(v, &s.Bold)
		case "dim":
			need = assignBool(v, &s.Dim)
		case "italic":
			need = assignBool(v, &s.Italic)
		case "underlined":
			need = assignBool(v, &s.Underlined)
		case "blink":
			need = assignBool(v, &s.Blink)
		case "inverse":
			need = assignBool(v, &s.Inverse)

		default:
			return fmt.Errorf("unrecognized option '%s'", k)
		}

		if need != "" {
			return fmt.Errorf("value to option '%s' must be a %s", k, need)
		}
	}

	return nil
}

func parseColor(name string) Color {
	if color, ok := colorByName[name]; ok {
		return color
	}
	if strings.HasPrefix(name, "color") {
		i, err := strconv.Atoi(name[5:])
		if err == nil && 0 <= i && i < 256 {
			return XTerm256Color(uint8(i))
		}
	} else if strings.HasPrefix(name, "#") && len(name) == 7 {
		r, rErr := strconv.ParseUint(name[1:3], 16, 8)
		g, gErr := strconv.ParseUint(name[3:5], 16, 8)
		b, bErr := strconv.ParseUint(name[5:7], 16, 8)
		if rErr == nil && gErr == nil && bErr == nil {
			return TrueColor(uint8(r), uint8(g), uint8(b))
		}
	}
	return nil
}

var sgrStyling = map[int]Styling{
	1:  Bold,
	2:  Dim,
	4:  Underlined,
	5:  Blink,
	7:  Inverse,
	30: FgBlack,
	31: FgRed,
	32: FgGreen,
	33: FgYellow,
	34: FgBlue,
	35: FgMagenta,
	36: FgCyan,
	37: FgWhite,
	40: BgBlack,
	41: BgRed,
	42: BgGreen,
	43: BgYellow,
	44: BgBlue,
	45: BgMagenta,
	46: BgCyan,
	47: BgWhite,
}

// StyleFromSGR builds a Style from a ECMA-48 Set Graphics Rendition sequence, .
// a semicolon-delimited string of valid attribute codes Invalid codes are     .
// ignored                                                                     .
func StyleFromSGR(s string) Style {
	style := Style{}
	for _, part := range strings.Split(s, ";") {
		code, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		if styling, ok := sgrStyling[code]; ok {
			styling.transform(&style)
		}
	}
	return style
}

package styled

import "fmt"

type Color uint

const (
	ColorDefault Color = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorLightGray
	ColorGray
	ColorLightRed
	ColorLightGreen
	ColorLightYellow
	ColorLightBlue
	ColorLightMagenta
	ColorLightCyan
	ColorWhite
)

func GetColorFromString(colString string) (Color, error) {
	switch colString {
	case "default":
		return ColorDefault, nil
	case "black":
		return ColorBlack, nil
	case "red":
		return ColorRed, nil
	case "green":
		return ColorGreen, nil
	case "yellow":
		return ColorYellow, nil
	case "blue":
		return ColorBlue, nil
	case "magenta":
		return ColorMagenta, nil
	case "cyan":
		return ColorCyan, nil
	case "lightgray":
		return ColorLightGray, nil
	case "gray":
		return ColorGray, nil
	case "lightred":
		return ColorLightRed, nil
	case "lightgreen":
		return ColorLightGreen, nil
	case "lightyellow":
		return ColorLightYellow, nil
	case "lightblue":
		return ColorLightBlue, nil
	case "lightmagenta":
		return ColorLightMagenta, nil
	case "lightcyan":
		return ColorLightCyan, nil
	case "white":
		return ColorWhite, nil
	}

	return 0, fmt.Errorf("color %s not recognized", colString)
}

func (c Color) String() string {
	switch c {
	case ColorDefault:
		return "default"
	case ColorBlack:
		return "black"
	case ColorRed:
		return "red"
	case ColorGreen:
		return "green"
	case ColorYellow:
		return "yellow"
	case ColorBlue:
		return "blue"
	case ColorMagenta:
		return "magenta"
	case ColorCyan:
		return "cyan"
	case ColorLightGray:
		return "lightgray"
	case ColorGray:
		return "gray"
	case ColorLightRed:
		return "lightred"
	case ColorLightGreen:
		return "lightgreen"
	case ColorLightYellow:
		return "lightyellow"
	case ColorLightBlue:
		return "lightblue"
	case ColorLightMagenta:
		return "lightmagenta"
	case ColorLightCyan:
		return "lightcyan"
	case ColorWhite:
		return "white"
	}

	return fmt.Sprintf("color %d", c)
}

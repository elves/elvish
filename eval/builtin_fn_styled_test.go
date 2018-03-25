package eval

import (
	"fmt"
	"testing"
)

// Taken from the original implementation and expanded with default foreground color.
var styleTranslationTable = map[string]string{
	"bold":       "1",
	"dim":        "2",
	"italic":     "3",
	"underlined": "4",
	"blink":      "5",
	"inverse":    "7",

	"default":      "39",
	"black":        "30",
	"red":          "31",
	"green":        "32",
	"yellow":       "33",
	"blue":         "34",
	"magenta":      "35",
	"cyan":         "36",
	"lightgray":    "37",
	"gray":         "90",
	"lightred":     "91",
	"lightgreen":   "92",
	"lightyellow":  "93",
	"lightblue":    "94",
	"lightmagenta": "95",
	"lightcyan":    "96",
	"white":        "97",

	"bg-default":      "49",
	"bg-black":        "40",
	"bg-red":          "41",
	"bg-green":        "42",
	"bg-yellow":       "43",
	"bg-blue":         "44",
	"bg-magenta":      "45",
	"bg-cyan":         "46",
	"bg-lightgray":    "47",
	"bg-gray":         "100",
	"bg-lightred":     "101",
	"bg-lightgreen":   "102",
	"bg-lightyellow":  "103",
	"bg-lightblue":    "104",
	"bg-lightmagenta": "105",
	"bg-lightcyan":    "106",
	"bg-white":        "107",
}

func TestCompatibility(t *testing.T) {
	const text = "abc"

	tests := make([]Test, len(styleTranslationTable))
	i := 0
	for k, v := range styleTranslationTable {
		tests[i] = That(fmt.Sprintf("print (styled %s %s)", text, k)).Prints(fmt.Sprintf("\033[%sm%s\033[m", v, text))
		i++
	}

	runTests(t, tests)
}

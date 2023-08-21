package ui

import (
	"testing"

	"src.elv.sh/pkg/testutil"
)

func TestStyleSGR(t *testing.T) {
	// Test the SGR sequences of style attributes indirectly via VTString of
	// Text, since that is how they are used.
	testTextVTString(t, []textVTStringTest{
		{T("foo", Bold), "\033[;1mfoo\033[m"},
		{T("foo", Dim), "\033[;2mfoo\033[m"},
		{T("foo", Italic), "\033[;3mfoo\033[m"},
		{T("foo", Underlined), "\033[;4mfoo\033[m"},
		{T("foo", Blink), "\033[;5mfoo\033[m"},
		{T("foo", Inverse), "\033[;7mfoo\033[m"},
		{T("foo", FgRed), "\033[;31mfoo\033[m"},
		{T("foo", BgRed), "\033[;41mfoo\033[m"},
		{T("foo", Bold, FgRed, BgBlue), "\033[;1;31;44mfoo\033[m"},
	})
}

func TestStyleSGR_NoColor(t *testing.T) {
	testutil.Set(t, &NoColor, true)
	testTextVTString(t, []textVTStringTest{
		{T("foo", FgRed), "\033[mfoo"},
		{T("foo", BgRed), "\033[mfoo"},
		{T("foo", FgRed, BgBlue), "\033[mfoo"},
	})
}

type mergeFromOptionsTest struct {
	style     Style
	options   map[string]any
	wantStyle Style
	wantErr   string
}

var mergeFromOptionsTests = []mergeFromOptionsTest{
	// Parsing of each possible key.
	kv("fg-color", "red", Style{Fg: Red}),
	kv("bg-color", "red", Style{Bg: Red}),
	kv("bold", true, Style{Bold: true}),
	kv("dim", true, Style{Dim: true}),
	kv("italic", true, Style{Italic: true}),
	kv("underlined", true, Style{Underlined: true}),
	kv("blink", true, Style{Blink: true}),
	kv("inverse", true, Style{Inverse: true}),
	// Merging with existing options.
	{
		style: Style{Bold: true, Dim: true},
		options: map[string]any{
			"bold": false, "fg-color": "red",
		},
		wantStyle: Style{Dim: true, Fg: Red},
	},
	// Bad key.
	{
		options: map[string]any{"bad": true},
		wantErr: "unrecognized option 'bad'",
	},
	// Bad type for color field.
	{
		options: map[string]any{"fg-color": true},
		wantErr: "value for option 'fg-color' must be a valid color string",
	},
	// Bad type for bool field.
	{
		options: map[string]any{"bold": ""},
		wantErr: "value for option 'bold' must be a bool value",
	},
}

// A helper for constructing a test case whose input is a single key-value pair.
func kv(k string, v any, s Style) mergeFromOptionsTest {
	return mergeFromOptionsTest{
		options: map[string]any{k: v}, wantStyle: s,
	}
}

func TestMergeFromOptions(t *testing.T) {
	for _, test := range mergeFromOptionsTests {
		style := test.style
		err := style.MergeFromOptions(test.options)
		if style != test.wantStyle {
			t.Errorf("(%v).MergeFromOptions(%v) -> %v, want %v",
				test.style, test.options, style, test.wantStyle)
		}
		if err == nil {
			if test.wantErr != "" {
				t.Errorf("got error nil, want %v", test.wantErr)
			}
		} else {
			if err.Error() != test.wantErr {
				t.Errorf("got error %v, want error with message %s", err, test.wantErr)
			}
		}
	}
}

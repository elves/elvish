// Styledown is a simple markup language for representing styled text.
//
// In the most basic form, Styledown markup consists of alternating text
// lines and style lines, where each character in the style line specifies
// the style of the character directly above it. For example:
//
//	foobar
//	***###
//	lorem
//	_____
//
// represents two lines:
//
//  1. "foo" in bold plus "bar" in reverse video
//  2. "lorem" in underline
//
// The following style characters are built-in:
//
//   - space for no style
//   - * for bold
//   - _ for underline
//   - # for reverse video
//
// This package can be used as a Go library or via Elvish's render-styledown
// command (https://elv.sh/ref/builtin.html#render-styledown).
//
// # Double-width characters
//
// Characters in text and style lines are matched up using their visual
// width, as calculated by [src.elv.sh/pkg/wcwidth.OfRune]. This means that
// double-width characters need to have their style character doubled:
//
//	好 foo
//	** ###
//
// The two style characters must be the same.
//
// # Configuration stanza
//
// An optional configuration stanza can follow the text and style lines (the
// content stanza), separated by a single newline. It can define additional
// style characters like this:
//
//	foobar
//	rrrGGG
//
//	r fg-red
//	G inverse fg-green
//
// Each line consists of the style character and one or more stylings as
// recognized by [src.elv.sh/pkg/ui.ParseStyling], separated by whitespaces. The
// character must be a single Unicode codepoint and have a visual width of 1.
//
// The configuration stanza can also contain additional options, and there's
// currently just one:
//
//   - no-eol: suppress the newline after the last line
//
// # Rationale
//
// Styledown is suitable for authoring a large chunk of styled text when the
// exact width and alignment of text need to be preserved.
//
// For example, it can be used to manually create and edit terminal mockups. In
// future it will be used in Elvish's tests for its terminal UI.
package styledown

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/wcwidth"
)

// Render renders Styledown markup. If the markup has parse errors, the error
// will start with "line x", where x is a 1-based line number.
func Render(s string) (ui.Text, error) {
	lines := strings.Split(s, "\n")
	i := 0
	for ; i+1 < len(lines) && wcwidth.Of(lines[i]) == wcwidth.Of(lines[i+1]); i += 2 {
	}
	contentLines := i
	if i < len(lines) {
		if lines[i] != "" {
			return nil, fmt.Errorf(
				"line %d: content and configuration stanzas must be separated by a newline", 1+i)
		}
		i++
	}
	opts, stylesheet, err := parseConfig(lines[i:], i+1)
	if err != nil {
		return nil, err
	}

	var tb ui.TextBuilder
	for i := 0; i < contentLines; i += 2 {
		if i > 0 {
			tb.WriteText(ui.T("\n"))
		}
		text, style := []rune(lines[i]), []rune(lines[i+1])
		for len(text) > 0 {
			r := text[0]
			w := wcwidth.OfRune(r)
			if !same(style[:w]) {
				return nil, fmt.Errorf(
					"line %d: inconsistent style %q for multi-width character %q",
					i+2, string(style[:w]), string(r))
			}
			styling, ok := stylesheet[style[0]]
			if !ok {
				return nil, fmt.Errorf(
					"line %d: unknown style %q", i+2, string(style[0]))
			}
			tb.WriteText(ui.T(string(r), styling))
			text = text[1:]
			style = style[w:]
		}
	}
	if !opts.noEOL {
		tb.WriteText(ui.T("\n"))
	}
	return tb.Text(), nil
}

type options struct {
	noEOL bool
}

func parseConfig(lines []string, firstLineNo int) (options, map[rune]ui.Styling, error) {
	var opts options
	stylesheet := map[rune]ui.Styling{
		' ': ui.Reset,
		'*': ui.Bold,
		'_': ui.Underlined,
		'#': ui.Inverse,
	}
	for i, line := range lines {
		if line == "" {
			continue
		}
		if line == "no-eol" {
			opts.noEOL = true
			continue
		}
		// Parse a style character definition.
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return options{}, nil, fmt.Errorf(
				"line %d: invalid configuration line", i+firstLineNo)
		}

		r, _ := utf8.DecodeRuneInString(fields[0])
		if string(r) != fields[0] {
			return options{}, nil, fmt.Errorf(
				"line %d: style character %q not a single character", i+firstLineNo, fields[0])
		}
		if wcwidth.OfRune(r) != 1 {
			return options{}, nil, fmt.Errorf(
				"line %d: style character %q not single-width", i+firstLineNo, fields[0])
		}

		stylingString := strings.Join(fields[1:], " ")
		styling := ui.ParseStyling(stylingString)
		if styling == nil {
			return options{}, nil, fmt.Errorf(
				"line %d: invalid styling string %q", i+firstLineNo, stylingString)
		}
		stylesheet[r] = styling
	}
	return opts, stylesheet, nil
}

func same[T comparable](s []T) bool {
	for i := 0; i+1 < len(s); i++ {
		if s[i] != s[i+1] {
			return false
		}
	}
	return true
}

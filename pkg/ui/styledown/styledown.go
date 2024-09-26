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
//  1. "foo" in bold, followed immediately by "bar" in reverse video
//  2. "lorem" in underline
//
// Note the following:
//
//   - Style characters like "*" are defined in [BuiltinStyleChars].
//   - A trailing newline after the last line is assumed.
//
// Both aspects can be altered by the configuration stanza (see below).
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
// The alternating text and style lines constitute the "content stanza". It can
// be followed by an optional configuration stanza, separated by a single
// newline.
//
// The configuration stanza can define additional style characters like this:
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
// Styledown is suitable for authoring large chunks of styled text. Its main
// advantage is that it preserves the alignment of text. Compare the following
// Styledown code:
//
//	foo:    100
//	###
//	foobar: 200
//	###___
//
// With the following hypothetical HTML-like format:
//
//	<i>foo</i>:    100
//	<i>foo</i><u>bar</u>: 200
//
// The Styledown example makes it clear that "100" and "200" are aligned. This
// property makes Styledown particularly suited for terminal mockups.
//
// Styldown is also used in Elvish's tests that need to represent styled text in
// a visual pure-text format.
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
			return nil, fmt.Errorf("line %d: text line must be matched by a style line", i+1)
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
			if w == 0 {
				return nil, fmt.Errorf("line %d: zero-width character is not allowed", i+1)
			}
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

// BuiltinStyleChars defines the styling characters that are recognized by
// default.
var BuiltinStyleChars = map[rune]ui.Styling{
	' ': ui.Reset,
	'*': ui.Bold,
	'_': ui.Underlined,
	'#': ui.Inverse,
}

func parseConfig(lines []string, firstLineNo int) (options, map[rune]ui.Styling, error) {
	var opts options
	stylesheet := map[rune]ui.Styling{}
	for i, line := range lines {
		if line == "" {
			continue
		}
		if line == "no-eol" {
			opts.noEOL = true
			continue
		}
		// Parse a style character definition.
		r, styling, err := parseStyleCharDef(line)
		if err != nil {
			return options{}, nil, fmt.Errorf("line %d: %w", i+firstLineNo, err)
		}
		if _, defined := stylesheet[r]; defined {
			return options{}, nil, fmt.Errorf(
				"line %d: duplicate style definition for %q", i+firstLineNo, r)
		}
		stylesheet[r] = styling
	}
	for r, st := range BuiltinStyleChars {
		if _, defined := stylesheet[r]; !defined {
			stylesheet[r] = st
		}
	}
	return opts, stylesheet, nil
}

func parseStyleCharDef(line string) (rune, ui.Styling, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0, nil, fmt.Errorf("invalid configuration line")
	}

	r, _ := utf8.DecodeRuneInString(fields[0])
	if string(r) != fields[0] {
		return 0, nil, fmt.Errorf("style character %q not a single character", fields[0])
	}
	if wcwidth.OfRune(r) != 1 {
		return 0, nil, fmt.Errorf("style character %q not single-width", fields[0])
	}

	stylingString := strings.Join(fields[1:], " ")
	styling := ui.ParseStyling(stylingString)
	if styling == nil {
		return 0, nil, fmt.Errorf("invalid styling string %q", stylingString)
	}
	return r, styling, nil
}

// Derender converts a Text to Styledown markup.
//
// The styleDefs string is used to supply style characters in addition to
// [BuiltinStyleChars]; it uses the same syntax as configuration lines in
// Styledown. Only those that are actually used will appear in the markup's
// configuration stanza, sorted by first use.
//
// This function returns an error if any of the following is true:
//
//   - styleDefs is invalid
//   - styleDefs contains two characters for the same style
//   - t contains segments with a style not covered by the map merged from
//     [BuiltinStyleChars] and styleDefs
func Derender(t ui.Text, styleDefs string) (string, error) {
	// Reverse map from Style to run.
	charForStyle := map[ui.Style]rune{}
	// Stores original style char definition lines. Also used later to track
	// whether a char definition has been written, by setting written values to
	// "".
	charDef := map[rune]string{}
	for i, line := range strings.Split(styleDefs, "\n") {
		if line == "" {
			continue
		}
		r, styling, err := parseStyleCharDef(line)
		if err != nil {
			return "", fmt.Errorf("styleDefs line %d: %w", i+1, err)
		}
		style := style(styling)
		if r2, ok := charForStyle[style]; ok {
			return "", fmt.Errorf("styleDefs line %d: %q defines the same style as %q", i+1, r, r2)
		}
		if _, defined := charDef[r]; defined {
			return "", fmt.Errorf("styleDefs line %d: %q is already defined", i+1, r)
		}
		charForStyle[style] = r
		charDef[r] = line
	}
	// Add builtin styles. We do this after parsing the supplied definitions
	// because they may override builtin styles, and we only want to add builtin
	// styles that were not overridden.
	for r, styling := range BuiltinStyleChars {
		if _, defined := charDef[r]; !defined {
			charForStyle[style(styling)] = r
		}
	}

	lines := t.SplitByRune('\n')

	var configStanza strings.Builder
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		// We have a trailing newline. Chop off the last element.
		lines = lines[:len(lines)-1]
	} else {
		configStanza.WriteString("no-eol\n")
	}

	var sb strings.Builder
	for i, line := range lines {
		var contentLine, styleLine strings.Builder
		for _, seg := range line {
			if r, ok := charForStyle[seg.Style]; ok {
				contentLine.WriteString(seg.Text)
				styleLine.WriteString(strings.Repeat(string(r), wcwidth.Of(seg.Text)))
				if charDef[r] != "" {
					configStanza.WriteString(charDef[r] + "\n")
					charDef[r] = ""
				}
			} else {
				return "", fmt.Errorf("line %d: style for segment %q has no char defined", i+1, seg.Text)
			}
		}
		sb.WriteString(contentLine.String() + "\n")
		sb.WriteString(styleLine.String() + "\n")
	}

	if configStanza.Len() > 0 {
		sb.WriteString("\n" + configStanza.String())
	}
	return sb.String(), nil
}

func style(s ui.Styling) ui.Style { return ui.ApplyStyling(ui.Style{}, s) }

func same[T comparable](s []T) bool {
	for i := 0; i+1 < len(s); i++ {
		if s[i] != s[i+1] {
			return false
		}
	}
	return true
}

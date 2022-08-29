package eval

import (
	"errors"
	"fmt"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

var errStyledSegmentArgType = errors.New("argument to styled-segment must be a string or a styled segment")

func init() {
	addBuiltinFns(map[string]any{
		"styled-segment": styledSegment,
		"styled":         styled,
	})
}

//elvdoc:fn styled-segment
//
// ```elvish
// styled-segment $object &fg-color=default &bg-color=default &bold=$false &dim=$false &italic=$false &underlined=$false &blink=$false &inverse=$false
// ```
//
// Constructs a styled segment, a building block for styled texts.
//
// - If `$object` is a string, constructs a styled segment with `$object` as the
//   content, and the properties specified by the options.
//
// - If `$object` is a styled segment, constructs a styled segment that is a
//   copy of `$object`, with the properties specified by the options overridden.
//
// The properties of styled segments can be inspected by indexing into it. Valid
// keys are the same as the options to `styled-segment`, plus `text` for the
// string content:
//
// ```elvish-transcript
// ~> var s = (styled-segment abc &bold)
// ~> put $s[text]
// ▶ abc
// ~> put $s[fg-color]
// ▶ default
// ~> put $s[bold]
// ▶ $true
// ```
//
// Prefer the high-level [`styled`](#styled) command to build and transform
// styled texts. Styled segments are a low-level construct, and you only have to
// deal with it when building custom style transformers.
//
// In the following example, a custom transformer sets the `inverse` property
// for every bold segment:
//
// ```elvish
// styled foo(styled bar bold) {|x| styled-segment $x &inverse=$x[bold] }
// # transforms "foo" + bold "bar" into "foo" + bold and inverse "bar"
// ```

// Turns a string or ui.Segment into a new ui.Segment with the attributes
// from the supplied options applied to it. If the input is already a Segment its
// attributes are copied and modified.
func styledSegment(options RawOptions, input any) (*ui.Segment, error) {
	var text string
	var style ui.Style

	switch input := input.(type) {
	case string:
		text = input
	case *ui.Segment:
		text = input.Text
		style = input.Style
	default:
		return nil, errStyledSegmentArgType
	}

	if err := style.MergeFromOptions(options); err != nil {
		return nil, err
	}

	return &ui.Segment{
		Text:  text,
		Style: style,
	}, nil
}

//elvdoc:fn styled
//
// ```elvish
// styled $object $style-transformer...
// ```
//
// Constructs a **styled text** by applying the supplied transformers to the
// supplied `$object`, which may be a string, a [styled
// segment](#styled-segment), or an existing styled text.
//
// Each `$style-transformer` can be one of the following:
//
// - A boolean attribute name:
//
//   - One of `bold`, `dim`, `italic`, `underlined`, `blink` and `inverse` for
//     setting the corresponding attribute.
//
//   - An attribute name prefixed by `no-` for unsetting the attribute.
//
//   - An attribute name prefixed by `toggle-` for toggling the attribute
//     between set and unset.
//
// - A color name for setting the text color, which may be one of the
//   following:
//
//   - One of the 8 basic ANSI colors: `black`, `red`, `green`, `yellow`,
//     `blue`, `magenta`, `cyan` and `white`.
//
//   - The bright variant of the 8 basic ANSI colors, with a `bright-` prefix.
//
//   - Any color from the xterm 256-color palette, as `colorX` (such as
//     `color12`).
//
//   - A 24-bit RGB color written as `#RRGGBB` (such as `'#778899'`).
//
//     **Note**: You need to quote such values, since an unquoted `#` introduces
//     a comment (e.g. use `'bg-#778899'` instead of `bg-#778899`).
//
// - A color name prefixed by `fg-` to set the foreground color. This has
//   the same effect as specifying the color name without the `fg-` prefix.
//
// - A color name prefixed by `bg-` to set the background color.
//
// - A function that receives a styled segment as the only argument and outputs
//   a single styled segment, which will be applied to all the segments.
//
// When a styled text is converted to a string the corresponding
// [ANSI SGR code](https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_.28Select_Graphic_Rendition.29_parameters)
// is built to render the style.
//
// Examples:
//
// ```elvish
// echo (styled foo red bold) # prints red bold "foo"
// echo (styled (styled foo red bold) green) # prints green bold "foo"
// ```
//
// A styled text can contain multiple [segments](#styled-segment) with different
// styles. Such styled texts can be constructed by concatenating multiple styled
// texts with the [compounding](language.html#compounding) syntax. Strings and
// styled segments are automatically "promoted" to styled texts when
// concatenating. Examples:
//
// ```elvish
// echo foo(styled bar red) # prints "foo" + red "bar"
// echo (styled foo bold)(styled bar red) # prints bold "foo" + red "bar"
// ```
//
// The individual segments in a styled text can be extracted by indexing:
//
// ```elvish
// var s = (styled abc red)(styled def green)
// put $s[0] $s[1]
// ```
//
// When printed to the terminal, a styled text is not affected by any existing
// SGR styles in effect, and it will always reset the SGR style afterwards. For
// example:
//
// ```elvish
// print "\e[1m"
// echo (styled foo red)
// echo bar
// # "foo" will be printed as red, but not bold
// # "bar" will be printed without any style
// ```

func styled(fm *Frame, input any, stylings ...any) (ui.Text, error) {
	var text ui.Text

	switch input := input.(type) {
	case string:
		text = ui.Text{&ui.Segment{
			Text:  input,
			Style: ui.Style{},
		}}
	case *ui.Segment:
		text = ui.Text{input.Clone()}
	case ui.Text:
		text = input.Clone()
	default:
		return nil, fmt.Errorf("expected string, styled segment or styled text; got %s", vals.Kind(input))
	}

	for _, styling := range stylings {
		switch styling := styling.(type) {
		case string:
			parsedStyling := ui.ParseStyling(styling)
			if parsedStyling == nil {
				return nil, fmt.Errorf("%s is not a valid style transformer", parse.Quote(styling))
			}
			text = ui.StyleText(text, parsedStyling)
		case Callable:
			for i, seg := range text {
				vs, err := fm.CaptureOutput(func(fm *Frame) error {
					return styling.Call(fm, []any{seg}, NoOpts)
				})
				if err != nil {
					return nil, err
				}

				if n := len(vs); n != 1 {
					return nil, fmt.Errorf("styling function must return a single segment; got %d values", n)
				} else if styledSegment, ok := vs[0].(*ui.Segment); !ok {
					return nil, fmt.Errorf("styling function must return a segment; got %s", vals.Kind(vs[0]))
				} else {
					text[i] = styledSegment
				}
			}

		default:
			return nil, fmt.Errorf("need string or callable; got %s", vals.Kind(styling))
		}
	}

	return text, nil
}

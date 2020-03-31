package eval

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/ui"
)

var errStyledSegmentArgType = errors.New("argument to styled-segment must be a string or a styled segment")

//elvdoc:fn styled-segment
//
// ```elvish
// styled-segment $object &fg-color=default &bg-color=default &bold=$false &dim=$false &italic=$false &underlined=$false &blink=$false &inverse=$false
// ```
//
// Constructs a styled segment and is a helper function for styled transformers.
// `$object` can be a plain string, a styled segment or a concatenation thereof.
// Probably the only reason to use it is to build custom style transformers:
//
// ```elvish
// fn my-awesome-style-transformer [seg]{ styled-segment $seg &bold=(not $seg[dim]) &dim=(not $seg[italic]) &italic=$seg[bold] }
// styled abc $my-awesome-style-transformer~
// ```
//
// As just seen the properties of styled segments can be inspected by indexing into
// it. Valid indices are the same as the options to `styled-segment` plus `text`.
//
// ```elvish
// s = (styled-segment abc &bold)
// put $s[text]
// put $s[fg-color]
// put $s[bold]
// ```

//elvdoc:fn styled
//
// ```elvish
// styled $object $style-transformer...
// ```
//
// Construct a styled text by applying the supplied transformers to the supplied
// object. `$object` can be either a string, a styled segment (see below), a styled
// text or an arbitrary concatenation of them. A `$style-transformer` is either:
//
// -   The name of a builtin style transformer, which may be one of the following:
//
// -   On of the attribute names `bold`, `dim`, `italic`, `underlined`, `blink`
// or `inverse` for setting the corresponding attribute
//
// -   An attribute name prefixed by `no-` for unsetting the attribute
//
// -   An attribute name prefixed by `toggle-` for toggling the attribute
// between set and unset
//
// -   A color name for setting the text color, which may be one of the
// following:
//
// -   One of the 8 basic ANSI colors: `black`, `red`, `green`, `yellow`,
// `blue`, `magenta`, `cyan` and `white`
//
// -   The bright variant of the 8 basic ANSI colors, with a `bright-`
// prefix
//
// -   Any color from the xterm 256-color palette, as `colorX` (such as
// `color12`)
//
// -   A 24-bit RGB color, as `#RRGGBB`, such as `#778899`.
//
// -   A color name prefixed by `bg-` to set the background color
//
// -   A lambda that receives a styled segment as the only argument and returns a
// single styled segment
//
// -   A function with the same properties as the lambda (provided via the
// `$transformer~` syntax)
//
// When a styled text is converted to a string the corresponding
// [ANSI SGR code](https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_.28Select_Graphic_Rendition.29_parameters)
// is built to render the style.
//
// A styled text is nothing more than a wrapper around a list of styled segments.
// They can be accessed by indexing into it.
//
// ```elvish
// s = (styled abc red)(styled def green)
// put $s[0] $s[1]
// ```

func init() {
	addBuiltinFns(map[string]interface{}{
		"styled-segment": styledSegment,
		"styled":         Styled,
	})
}

// Turns a string or ui.Segment into a new ui.Segment with the attributes
// from the supplied options applied to it. If the input is already a Segment its
// attributes are copied and modified.
func styledSegment(options RawOptions, input interface{}) (*ui.Segment, error) {
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

// Styled turns a string, a ui.Segment or a ui.Text into a ui.Text by applying
// the given stylings.
func Styled(fm *Frame, input interface{}, stylings ...interface{}) (ui.Text, error) {
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
				vs, err := fm.CaptureOutput(styling, []interface{}{seg}, NoOpts)
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

package eval

import (
	"errors"
	"fmt"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/ui/styledown"
)

var errStyledSegmentArgType = errors.New("argument to styled-segment must be a string or a styled segment")

func init() {
	addBuiltinFns(map[string]any{
		"styled-segment":   styledSegment,
		"styled":           styled,
		"render-styledown": styledown.Render,
	})
}

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

func styled(fm *Frame, input any, stylings ...any) (ui.Text, error) {
	var text ui.Text

	switch input := input.(type) {
	case string:
		text = ui.T(input)
	case *ui.Segment:
		text = ui.TextFromSegment(input)
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

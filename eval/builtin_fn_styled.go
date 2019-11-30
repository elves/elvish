package eval

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/ui"
)

var errStyledSegmentArgType = errors.New("argument to styled-segment must be a string or a styled segment")

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

	if err := style.ImportFromOptions(options); err != nil {
		return nil, err
	}

	return &ui.Segment{
		Text:  text,
		Style: style,
	}, nil
}

// Styled turns a string, a ui.Segment or a ui.Text into a ui.Text.
// This is done by applying a range of transformers to the input.
func Styled(fm *Frame, input interface{}, transformers ...interface{}) (ui.Text, error) {
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

	for _, transformer := range transformers {
		switch transformer := transformer.(type) {
		case string:
			transformerFn := ui.FindTransformer(transformer)
			if transformerFn == nil {
				return nil, fmt.Errorf("%s is not a valid style transformer", parse.Quote(transformer))
			}
			for _, seg := range text {
				transformerFn(&seg.Style)
			}
		case Callable:
			for i, seg := range text {
				vs, err := fm.CaptureOutput(transformer, []interface{}{seg}, NoOpts)
				if err != nil {
					return nil, err
				}

				if n := len(vs); n != 1 {
					return nil, fmt.Errorf("style transformers must return a single styled segment; got %d values", n)
				} else if transformedSegment, ok := vs[0].(*ui.Segment); !ok {
					return nil, fmt.Errorf("style transformers must return a styled segment; got %s", vals.Kind(vs[0]))
				} else {
					text[i] = transformedSegment
				}
			}

		default:
			return nil, fmt.Errorf("need string or callable; got %s", vals.Kind(transformer))
		}
	}

	return text, nil
}

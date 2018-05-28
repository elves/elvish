package eval

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/styled"
)

var errStyledSegmentArgType = errors.New("argument to styled-segment must be a string or a styled segment")

func init() {
	addBuiltinFns(map[string]interface{}{
		"styled-segment": styledSegment,
		"styled":         Styled,
	})
}

// Turns a string or styled Segment into a new styled Segment with the attributes
// from the supplied options applied to it. If the input is already a Segment its
// attributes are copied and modified.
func styledSegment(options RawOptions, input interface{}) (*styled.Segment, error) {
	var text string
	var style styled.Style

	switch input := input.(type) {
	case string:
		text = input
	case *styled.Segment:
		text = input.Text
		style = input.Style
	default:
		return nil, errStyledSegmentArgType
	}

	if err := style.ImportFromOptions(options); err != nil {
		return nil, err
	}

	return &styled.Segment{
		Text:  text,
		Style: style,
	}, nil
}

// Styled turns a string, a styled Segment or a styled Text into a styled Text.
// This is done by applying a range of transformers to the input.
func Styled(fm *Frame, input interface{}, transformers ...interface{}) (*styled.Text, error) {
	var text styled.Text

	switch input := input.(type) {
	case string:
		text = styled.Text{styled.Segment{
			Text:  input,
			Style: styled.Style{},
		}}
	case *styled.Segment:
		text = styled.Text{*input}
	case *styled.Text:
		text = *input
	default:
		return nil, fmt.Errorf("expected string, styled segment or styled text; got %s", vals.Kind(input))
	}

	for _, transformer := range transformers {
		switch transformer := transformer.(type) {
		case string:
			transformerFn := styled.FindTransformer(transformer)
			if transformerFn == nil {
				return nil, fmt.Errorf("'%s' is no valid style transformer", transformer)
			}

			for i, segment := range text {
				text[i] = transformerFn(segment)
			}

		case Callable:
			for i, segment := range text {
				vs, err := fm.CaptureOutput(transformer, []interface{}{&segment}, NoOpts)
				if err != nil {
					return nil, err
				}

				if n := len(vs); n != 1 {
					return nil, fmt.Errorf("style transformers must return a single styled segment; got %d", n)
				} else if transformedSegment, ok := vs[0].(*styled.Segment); !ok {
					return nil, fmt.Errorf("style transformers must return a styled segment; got %s", vals.Kind(vs[0]))
				} else {
					text[i] = *transformedSegment
				}
			}

		default:
			return nil, fmt.Errorf("need string or callable; got %s", vals.Kind(transformer))
		}
	}

	return &text, nil
}

package eval

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/styled"
	"github.com/xiaq/persistent/hashmap"
)

var styledTransformers hashmap.Map

func init() {
	addBuiltinFns(map[string]interface{}{
		"styled-segment": styledSegment,
		"styled":         StyledBuiltin,
	})

	transformers := make(map[interface{}]interface{})
	for k, v := range styled.SegmentTransformers {
		transformers[k] = NewBuiltinFn("style-transform "+k, v)
	}

	styledTransformers = vals.MakeMap(transformers)
	builtinNs.Add("styled-transformers", vars.FromPtr(&styledTransformers))
}

// Turns a string or styled Segment into a new styled Segment with the attributes
// from the supplied options applied to it. If the input is already a Segment its
// attributes are copied and modified.
func styledSegment(options RawOptions, input interface{}) (*styled.Segment, error) {
	switch input := input.(type) {
	case string:
		fg, err := styled.ForegroundColorFromOptions(options)
		if err != nil {
			return nil, err
		}

		bg, err := styled.BackgroundColorFromOptions(options)
		if err != nil {
			return nil, err
		}

		textStyle, err := styled.TextStyleFromMap(options)
		if err != nil {
			return nil, err
		}

		style := styled.Style{
			Foreground: fg,
			Background: bg,
			TextStyle:  *textStyle,
		}

		return &styled.Segment{
			Text:  input,
			Style: style,
		}, nil
	case styled.Segment:
		fg, err := styled.ForegroundColorFromOptions(options)
		if err != nil {
			return nil, err
		}
		if fg == nil {
			fg = input.Foreground
		}

		bg, err := styled.BackgroundColorFromOptions(options)
		if err != nil {
			return nil, err
		}
		if bg == nil {
			bg = input.Background
		}

		textStyle, err := styled.TextStyleFromMap(options)
		if err != nil {
			return nil, err
		}

		style := styled.Style{
			Foreground: fg,
			Background: bg,
			TextStyle:  input.TextStyle.Merge(textStyle),
		}

		return &styled.Segment{
			Text:  input.Text,
			Style: style,
		}, nil
	}

	return nil, errors.New("expected string or styled segment")
}

// Turns a string, a styled Segment or a styled Text into a styled Text. This is done by
// applying a range of transformers to the input.
func StyledBuiltin(fm *Frame, input interface{}, transformers ...interface{}) (*styled.Text, error) {
	var text styled.Text

	switch input := input.(type) {
	case string:
		text = styled.Text{styled.Segment{Text: input}}
	case styled.Segment:
		text = styled.Text{input}
	case styled.Text:
		text = input
	default:
		return nil, fmt.Errorf("expected string, styled segment or styled text; got %s", vals.Kind(input))
	}

	for _, transformer := range transformers {
		var fn Callable
		switch transformer := transformer.(type) {
		case string:
			if transformerFn, ok := styledTransformers.Index(transformer); ok {
				if transformerFn, ok := transformerFn.(Callable); ok {
					fn = transformerFn
				} else {
					return nil, fmt.Errorf("transformer %s is not callable", transformer)
				}
			} else {
				return nil, fmt.Errorf("transformer %s not found in $styled-transformers", transformer)
			}
		case Callable:
			fn = transformer
		default:
			return nil, fmt.Errorf("need string or callable; got %s", vals.Kind(transformer))
		}

		for i, segment := range text {
			vs, err := fm.CaptureOutput(fn, []interface{}{segment}, NoOpts)
			if err != nil {
				return nil, err
			}

			if n := len(vs); n != 1 {
				return nil, fmt.Errorf("style transformers must return a single styled segment; got %d", n)
			}

			switch transformedSegment := vs[0].(type) {
			case styled.Segment:
				text[i] = transformedSegment
			case *styled.Segment:
				text[i] = *transformedSegment
			default:
				return nil, fmt.Errorf("style transformers must return a styled segment; got %s", vals.Kind(vs[0]))
			}
		}
	}

	return &text, nil
}

package eval

import (
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/styled"
	"github.com/xiaq/persistent/hashmap"
)

var styledTransformers hashmap.Map

func init() {
	addBuiltinFns(map[string]interface{}{
		"styled-segment": styledSegment,
		"styled":         styledBuiltin,
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
func styledSegment(options RawOptions, input interface{}) styled.Segment {
	switch input := input.(type) {
	case string:
		style := styled.Style{
			Foreground: styled.ForegroundColorFromOptions(options),
			Background: styled.BackgroundColorFromOptions(options),
			TextStyle:  styled.TextStyleFromMap(options),
		}

		return styled.Segment{
			Text:  input,
			Style: style,
		}
	case styled.Segment:
		fg := styled.ForegroundColorFromOptions(options)
		bg := styled.BackgroundColorFromOptions(options)
		if fg == nil {
			fg = input.Foreground
		}
		if bg == nil {
			bg = input.Background
		}
		style := styled.Style{
			Foreground: fg,
			Background: bg,
			TextStyle:  input.TextStyle.Merge(styled.TextStyleFromMap(options)),
		}

		return styled.Segment{
			Text:  input.Text,
			Style: style,
		}
	default:
		throwf("expected string or styled segment")
		panic("unreachable")
	}
}

// Turns a string, a styled Segment or a styled Text into a styled Text. This is done by
// applying a range of transformers to the input.
func styledBuiltin(fm *Frame, input interface{}, transformers ...interface{}) styled.Text {
	var text styled.Text

	switch input := input.(type) {
	case string:
		text = styled.Text{styled.Segment{Text: input}}
	case styled.Segment:
		text = styled.Text{input}
	case styled.Text:
		text = input
	default:
		throwf("expected string, styled segment or styled text; got %s", vals.Kind(input))
	}

	for _, transformer := range transformers {
		var fn Callable
		switch transformer := transformer.(type) {
		case string:
			if transformerFn, ok := styledTransformers.Index(transformer); ok {
				if transformerFn, ok := transformerFn.(Callable); ok {
					fn = transformerFn
				} else {
					throwf("transformer %s is not callable", transformer)
				}
			} else {
				throwf("transformer %s not found in $styled-transformers", transformer)
			}
		case Callable:
			fn = transformer
		default:
			throwf("need string or callable; got %s", vals.Kind(transformer))
		}

		for i, segment := range text {
			vs, err := fm.CaptureOutput(fn, []interface{}{segment}, NoOpts)
			maybeThrow(err)

			throw := false
			if n := len(vs); n != 1 {
				throw = true
			} else if transformedSegment, ok := vs[0].(styled.Segment); !ok {
				throw = true
			} else {
				text[i] = transformedSegment
			}

			if throw {
				throwf("style transformers are expected to return exactly one styled segment")
			}
		}
	}

	return text
}

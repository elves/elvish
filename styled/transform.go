package styled

import (
	"fmt"
	"strings"
)

func FindTransformer(transformerName string) (func(Segment) Segment, error) {
	var innerTransformer func(*Segment)

	// Special cases for handling inverse
	switch {
	case transformerName == "inverse":
		innerTransformer = func(s *Segment) { s.Inverse = !s.Inverse }
	case transformerName == "force-inverse":
		innerTransformer = func(s *Segment) { s.Inverse = true }
	case transformerName == "force-no-inverse":
		innerTransformer = func(s *Segment) { s.Inverse = false }

	case strings.HasPrefix(transformerName, "bg-"):
		innerTransformer = buildColorTransformer(strings.TrimPrefix(transformerName, "bg-"), false)
	case strings.HasPrefix(transformerName, "no-"):
		innerTransformer = buildBoolTransformer(strings.TrimPrefix(transformerName, "no-"), false)

	default:
		innerTransformer = buildColorTransformer(transformerName, true)

		if innerTransformer == nil {
			innerTransformer = buildBoolTransformer(transformerName, true)
		}
	}

	if innerTransformer == nil {
		return nil, fmt.Errorf("'%s' is no valid style transformer", transformerName)
	}

	return func(segment Segment) Segment {
		innerTransformer(&segment)
		return segment
	}, nil
}

func buildColorTransformer(transformerName string, setForeground bool) func(*Segment) {
	if isValidColorName(transformerName) {
		if transformerName == "default" {
			transformerName = ""
		}

		if setForeground {
			return func(s *Segment) {
				s.Foreground = transformerName
			}
		} else {
			return func(s *Segment) {
				s.Background = transformerName
			}
		}
	}

	return nil
}

func buildBoolTransformer(transformerName string, val bool) func(*Segment) {
	switch transformerName {
	case "bold":
		return func(s *Segment) { s.Bold = val }
	case "dim":
		return func(s *Segment) { s.Dim = val }
	case "italic":
		return func(s *Segment) { s.Italic = val }
	case "underlined":
		return func(s *Segment) { s.Underlined = val }
	case "blink":
		return func(s *Segment) { s.Blink = val }
	}

	return nil
}

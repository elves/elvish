package styled

type segmentTransformer func(Segment) Segment

var SegmentTransformers map[string]segmentTransformer

func init() {
	SegmentTransformers = make(map[string]segmentTransformer)

	makeBool := func(assign func(*Segment)) segmentTransformer {
		return func(segment Segment) Segment {
			assign(&segment)
			return segment
		}
	}

	SegmentTransformers["bold"] = makeBool(func(segment *Segment) { segment.Bold = true })
	SegmentTransformers["dim"] = makeBool(func(segment *Segment) { segment.Dim = true })
	SegmentTransformers["italic"] = makeBool(func(segment *Segment) { segment.Italic = true })
	SegmentTransformers["underlined"] = makeBool(func(segment *Segment) { segment.Underlined = true })
	SegmentTransformers["blink"] = makeBool(func(segment *Segment) { segment.Blink = true })
	SegmentTransformers["inverse"] = makeBool(func(segment *Segment) { segment.Inverse = !segment.Inverse })

	makeFg := func(col string) segmentTransformer {
		return func(segment Segment) Segment {
			segment.Foreground = col
			return segment
		}
	}
	makeBg := func(col string) segmentTransformer {
		return func(segment Segment) Segment {
			segment.Background = col
			return segment
		}
	}

	colors := []string{
		"default",
		"black",
		"red",
		"green",
		"yellow",
		"blue",
		"magenta",
		"cyan",
		"lightgray",
		"gray",
		"lightred",
		"lightgreen",
		"lightyellow",
		"lightblue",
		"lightmagenta",
		"lightcyan",
		"white",
	}
	for _, col := range colors {
		SegmentTransformers[col] = makeFg(col)
		SegmentTransformers["bg-"+col] = makeBg(col)
	}
}

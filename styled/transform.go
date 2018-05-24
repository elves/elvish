package styled

type segmentTransformer func(Segment) Segment

var SegmentTransformers map[string]segmentTransformer

func init() {
	myTrue := true

	SegmentTransformers = map[string]segmentTransformer{
		"bold": func(segment Segment) Segment {
			segment.bold = &myTrue
			return segment
		},
		"dim": func(segment Segment) Segment {
			segment.dim = &myTrue
			return segment
		},
		"italic": func(segment Segment) Segment {
			segment.italic = &myTrue
			return segment
		},
		"underlined": func(segment Segment) Segment {
			segment.underlined = &myTrue
			return segment
		},
		"blink": func(segment Segment) Segment {
			segment.blink = &myTrue
			return segment
		},
		"inverse": func(segment Segment) Segment {
			var val bool
			if segment.inverse == nil || !(*segment.inverse) {
				val = true
			}
			segment.inverse = &val
			return segment
		},
	}

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

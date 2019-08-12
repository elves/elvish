package ui

func writeHorizontalScrollbar(bb *BufferBuilder, n, low, high, width int) {
	slow, shigh := findScrollInterval(n, low, high, width)
	for i := 0; i < width; i++ {
		if slow <= i && i < shigh {
			bb.WriteRuneSGR(' ', styleForScrollBarThumb.String())
		} else {
			bb.WriteRuneSGR('━', styleForScrollBarArea.String())
		}
	}
}

func renderVerticalScrollbar(n, low, high, height int) *Buffer {
	slow, shigh := findScrollInterval(n, low, high, height)
	bb := NewBufferBuilder(1)
	for i := 0; i < height; i++ {
		if i > 0 {
			bb.Newline()
		}
		if slow <= i && i < shigh {
			bb.WriteRuneSGR(' ', styleForScrollBarThumb.String())
		} else {
			bb.WriteRuneSGR('│', styleForScrollBarArea.String())
		}
	}
	return bb.Buffer()
}

func findScrollInterval(n, low, high, height int) (int, int) {
	f := func(i int) int {
		return int(float64(i)/float64(n)*float64(height) + 0.5)
	}
	scrollLow, scrollHigh := f(low), f(high)
	if scrollLow == scrollHigh {
		if scrollHigh == high {
			scrollLow--
		} else {
			scrollHigh++
		}
	}
	return scrollLow, scrollHigh
}

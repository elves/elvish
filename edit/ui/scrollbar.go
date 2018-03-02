package ui

func writeHorizontalScrollbar(b *Buffer, n, low, high, width int) {
	slow, shigh := findScrollInterval(n, low, high, width)
	for i := 0; i < width; i++ {
		if slow <= i && i < shigh {
			b.Write(' ', styleForScrollBarThumb.String())
		} else {
			b.Write('━', styleForScrollBarArea.String())
		}
	}
}

func renderScrollbar(n, low, high, height int) *Buffer {
	slow, shigh := findScrollInterval(n, low, high, height)
	// Logger.Printf("low = %d, high = %d, n = %d, slow = %d, shigh = %d", low, high, n, slow, shigh)
	b := NewBuffer(1)
	for i := 0; i < height; i++ {
		if i > 0 {
			b.Newline()
		}
		if slow <= i && i < shigh {
			b.Write(' ', styleForScrollBarThumb.String())
		} else {
			b.Write('│', styleForScrollBarArea.String())
		}
	}
	return b
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

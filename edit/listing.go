package edit

import "unicode/utf8"

// listing encapsulates functionalities common for listing modes.
type listing struct {
	selected int
	filter   string
	onfilter func(string)
}

func (l *listing) modeLine(title string, width int) *buffer {
	// TODO keep it one line.
	b := newBuffer(width)
	b.writes(TrimWcWidth(title, width), styleForMode)
	b.writes(" ", "")
	b.writes(l.filter, styleForFilter)
	b.dot = b.cursor()
	return b
}

func (l *listing) list(get func(int) string, n, width, maxHeight int) *buffer {
	b := newBuffer(width)
	if n == 0 {
		b.writes(TrimWcWidth("(no result)", width), "")
		return b
	}
	low, high := findWindow(n, l.selected, maxHeight)
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		style := ""
		if i == l.selected {
			style = styleForSelected
		}
		b.writes(TrimWcWidth(get(i), width), style)
	}
	return b
}

func (l *listing) changeFilter(newfilter string) {
	l.filter = newfilter
	if l.onfilter != nil {
		l.onfilter(newfilter)
	}
}

func (l *listing) backspace() bool {
	_, size := utf8.DecodeLastRuneInString(l.filter)
	if size > 0 {
		l.changeFilter(l.filter[:len(l.filter)-size])
		return true
	}
	return false
}

func (l *listing) handleFilterKey(k Key) bool {
	if likeChar(k) {
		l.changeFilter(l.filter + string(k.Rune))
		return true
	}
	return false
}

func (l *listing) prev(cycle bool, n int) {
	if n == 0 {
		return
	}
	l.selected--
	if l.selected == -1 {
		if cycle {
			l.selected += n
		} else {
			l.selected++
		}
	}
}

func (l *listing) next(cycle bool, n int) {
	if n == 0 {
		return
	}
	l.selected++
	if l.selected == n {
		if cycle {
			l.selected -= n
		} else {
			l.selected--
		}
	}
}

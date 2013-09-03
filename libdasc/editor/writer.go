package editor

import (
	"os"
	"./tty"
)

type writer struct {
	line, col int
	width int
}

func newWriter() *writer {
	return &writer{width: int(tty.GetWinsize(0).Col)}
}

func (w *writer) write(r rune) error {
	var s string
	wd := wcwidth(r)
	if w.col + wd > w.width {
		// overflow
		w.line++
		w.col = wd
		s = "\n" + string(r)
	} else if w.col + wd == w.width {
		// just fit
		w.line++
		w.col = 0
		s = string(r) + "\n"
	} else {
		w.col++
		s = string(r)
	}
	_, err := os.Stdout.WriteString(s)
	return err
}

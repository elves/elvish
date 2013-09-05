package editor

import (
	"os"
	"./tty"
)

type writer struct {
	line, col int
	width int
	file *os.File
}

func newWriter(file *os.File) *writer {
	return &writer{
		width: int(tty.GetWinsize(int(file.Fd())).Col),
		file: file,
	}
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
	_, err := w.file.WriteString(s)
	return err
}

package edit

import (
	"os"
	"fmt"
	"bytes"
	"./tty"
	"../parse"
)

// cell is an indivisible unit on the screen. It is not necessarily 1 column
// wide.
type cell struct {
	rune
	width byte
	attr string
}

// buffer keeps the status of the screen that the line editor is concerned
// with. It usually reflects the last n lines of the screen.
type buffer [][]cell

// writer is the part of an Editor responsible for keeping the status of and
// updating the screen.
type writer struct {
	oldBuf, buf buffer
	line, col, width, indent int
	currentAttr string
}

func newWriter() *writer {
	writer := &writer{oldBuf: buffer{[]cell{}}}
	return writer
}

func (w *writer) startBuffer(file *os.File) {
	w.line = 0
	w.col = 0
	w.width = int(tty.GetWinsize(int(file.Fd())).Col)
	w.buf = [][]cell{make([]cell, w.width)}
	w.currentAttr = ""
}

func (w *writer) commitBuffer(file *os.File) error {
	bytesBuf := new(bytes.Buffer)

	newlines := len(w.oldBuf) - 1
	if newlines > 0 {
		fmt.Fprintf(bytesBuf, "\033[%dA", newlines)
	}
	bytesBuf.WriteString("\r\033[J")

	attr := ""
	for _, line := range w.buf {
		for _, c := range line {
			if c.width > 0 && c.attr != attr {
				fmt.Fprintf(bytesBuf, "\033[%sm", c.attr)
				attr = c.attr
			}
			bytesBuf.WriteString(string(c.rune))
		}
	}
	if attr != "" {
		bytesBuf.WriteString("\033[m")
	}
	_, err := file.Write(bytesBuf.Bytes())
	if err != nil {
		return err
	}

	w.oldBuf = w.buf
	return nil
}

func (w *writer) appendToLine(c cell) {
	w.buf[w.line] = append(w.buf[w.line], c)
	w.col += int(c.width)
}

func (w *writer) newline() {
	w.buf[w.line] = append(w.buf[w.line], cell{rune: '\n'})
	w.buf = append(w.buf, make([]cell, w.width))
	w.line++
	w.col = 0
	if w.indent > 0 {
		for i := 0; i < w.indent; i++ {
			w.appendToLine(cell{rune: ' ', width: 1})
		}
	}
}

func (w *writer) write(r rune) {
	wd := wcwidth(r)
	c := cell{r, byte(wd), w.currentAttr}

	if w.col + wd > w.width {
		w.newline()
		w.appendToLine(c)
	} else if w.col + wd == w.width {
		w.appendToLine(c)
		w.newline()
	} else {
		w.appendToLine(c)
	}
}

func (w *writer) refresh(prompt, text string, file *os.File) error {
	w.startBuffer(file)

	for _, r := range prompt {
		w.write(r)
	}

	if w.col * 2 < w.width {
		w.indent = w.col
	}

	l := parse.Lex("<interactive code>", text)

	for {
		token := l.NextItem()
		if token.Typ == parse.ItemEOF {
			break
		}
		w.currentAttr = attrForType[token.Typ]
		for _, r := range token.Val {
			w.write(r)
		}
	}

	return w.commitBuffer(file)
}

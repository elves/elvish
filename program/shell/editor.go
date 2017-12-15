package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type editor interface {
	ReadLine() (string, error)
	Close()
}

type minEditor struct {
	in  *bufio.Reader
	out io.Writer
}

func newMinEditor(in, out *os.File) *minEditor {
	return &minEditor{bufio.NewReader(in), out}
}

func (ed *minEditor) ReadLine() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		wd = "?"
	}
	fmt.Fprintf(ed.out, "%s> ", wd)
	line, err := ed.in.ReadString('\n')
	// Chop off the trailing \r on Windows.
	line = strings.TrimRight(line, "\r\n")
	return line, err
}

func (editor *minEditor) Close() {
}

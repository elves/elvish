package shell

import (
	"bufio"
	"io"
	"os"
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
	return ed.in.ReadString('\n')
}

func (editor *minEditor) Close() {
}

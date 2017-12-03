package shell

import (
	"bufio"
	"io"
	"os"
)

type minEditor struct {
	in  *bufio.Reader
	out io.Writer
}

func makeEditor(in, out *os.File, _, _ interface{}) *minEditor {
	return &minEditor{bufio.NewReader(in), out}
}

func (ed *minEditor) ReadLine() (string, error) {
	return ed.in.ReadString('\n')
}

func (editor *minEditor) Close() {
}

func handleSignal(_ os.Signal) {
}

package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/strutil"
)

// This type is the interface that the line editor has to satisfy. It is needed so that this package
// does not depend on the edit package.
type editor interface {
	ReadCode() (string, error)
	RunAfterCommandHooks(src parse.Source, duration float64, err error)
}

type minEditor struct {
	in  *bufio.Reader
	out io.Writer
}

func newMinEditor(in, out *os.File) *minEditor {
	return &minEditor{bufio.NewReader(in), out}
}

// RunAfterCommandHooks is a no-op in the minimum editor since it doesn't support
// `edit:after-command` hooks. The method is needed to satisfy the `editor` interface.
func (ed *minEditor) RunAfterCommandHooks(src parse.Source, duration float64, err error) {
}

func (ed *minEditor) ReadCode() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		wd = "?"
	}
	fmt.Fprintf(ed.out, "%s> ", wd)
	line, err := ed.in.ReadString('\n')
	return strutil.ChopLineEnding(line), err
}

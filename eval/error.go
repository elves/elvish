package eval

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/elves/elvish/parse"
)

// Error represents runtime errors in elvish constructs.
type Error struct {
	Inner error
}

func (Error) Kind() string {
	return "error"
}

func (e Error) Repr(indent int) string {
	if e.Inner == nil {
		return "$ok"
	}
	if r, ok := e.Inner.(Reprer); ok {
		return r.Repr(indent)
	}
	return "?(error " + parse.Quote(e.Inner.Error()) + ")"
}

func (e Error) Bool() bool {
	return e.Inner == nil
}

// OK is an alias for the zero value of Error.
var OK = Error{nil}

// multiError is multiple errors packed into one. It is used for reporting
// errors of pipelines, in which multiple forms may error.
type MultiError struct {
	Errors []Error
}

func (me MultiError) Repr(indent int) string {
	// TODO Make a more generalized ListReprBuilder and use it here.
	b := new(bytes.Buffer)
	b.WriteString("?(multi-error")
	elemIndent := indent + len("?(multi-error ")
	for _, e := range me.Errors {
		if indent > 0 {
			b.WriteString("\n" + strings.Repeat(" ", elemIndent))
		} else {
			b.WriteString(" ")
		}
		b.WriteString(e.Repr(elemIndent))
	}
	b.WriteString(")")
	return b.String()
}

func (me MultiError) Error() string {
	b := new(bytes.Buffer)
	b.WriteString("(")
	for i, e := range me.Errors {
		if i > 0 {
			b.WriteString(" | ")
		}
		if e.Inner == nil {
			b.WriteString("<nil>")
		} else {
			b.WriteString(e.Inner.Error())
		}
	}
	b.WriteString(")")
	return b.String()
}

func newMultiError(es ...Error) Error {
	return Error{MultiError{es}}
}

// Flow is a special type of Error used for control flows.
type flow uint

// Control flows.
const (
	Return flow = iota
	Break
	Continue
)

var flowNames = [...]string{
	"return", "break", "continue",
}

func (f flow) Repr(int) string {
	return "?(" + f.Error() + ")"
}

func (f flow) Error() string {
	if f >= flow(len(flowNames)) {
		return fmt.Sprintf("!(BAD FLOW: %v)", f)
	}
	return flowNames[f]
}

// ExternalCmdExit contains the exit status of external commands. If the
// command was stopped rather than terminated, the Pid field contains the pid
// of the process.
type ExternalCmdExit struct {
	syscall.WaitStatus
	CmdName string
	Pid     int
}

func NewExternalCmdExit(name string, ws syscall.WaitStatus, pid int) error {
	if ws.Exited() && ws.ExitStatus() == 0 {
		return nil
	}
	if !ws.Stopped() {
		pid = 0
	}
	return ExternalCmdExit{ws, name, pid}
}

func FakeExternalCmdExit(name string, exit int, sig syscall.Signal) ExternalCmdExit {
	return ExternalCmdExit{syscall.WaitStatus(exit<<8 + int(sig)), name, 0}
}

func (exit ExternalCmdExit) Error() string {
	ws := exit.WaitStatus
	quotedName := parse.Quote(exit.CmdName)
	switch {
	case ws.Exited():
		return quotedName + " exited with " + strconv.Itoa(ws.ExitStatus())
	case ws.Signaled():
		msg := quotedName + " killed by signal " + ws.Signal().String()
		if ws.CoreDump() {
			msg += " (core dumped)"
		}
		return msg
	case ws.Stopped():
		msg := quotedName + " stopped by signal " + fmt.Sprintf("%s (pid=%d)", ws.StopSignal(), exit.Pid)
		trap := ws.TrapCause()
		if trap != -1 {
			msg += fmt.Sprintf(" (trapped %v)", trap)
		}
		return msg
	default:
		return fmt.Sprint(quotedName, " has unknown WaitStatus ", ws)
	}
}

func allok(es []Error) bool {
	for _, e := range es {
		if e.Inner != nil {
			return false
		}
	}
	return true
}

// PprintError pretty prints an error. It understands specialized error types
// defined in this package.
func PprintError(e error) {
	switch e := e.(type) {
	case nil:
		fmt.Print("\033[32mok\033[m")
	case MultiError:
		fmt.Print("(")
		for i, c := range e.Errors {
			if i > 0 {
				fmt.Print(" | ")
			}
			PprintError(c.Inner)
		}
		fmt.Print(")")
	case flow:
		fmt.Print("\033[33m" + e.Error() + "\033[m")
	default:
		fmt.Print("\033[31;1m" + e.Error() + "\033[m")
	}
}

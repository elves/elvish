package eval

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/xiaq/persistent/hash"
)

// Exception represents an elvish exception. It is both a Value accessible to
// elvishscript, and the type of error returned by public facing evaluation
// methods like (*Evaler)PEval.
type Exception struct {
	Reason     error
	StackTrace *stackTrace
}

// A stack trace as a linked list of diag.Context. The head is the innermost
// stack. Since pipelines can call multiple functions in parallel, all the
// stackTrace nodes form a DAG.
type stackTrace struct {
	head *diag.Context
	next *stackTrace
}

// Cause returns the Cause field if err is an *Exception. Otherwise it returns
// err itself.
func Cause(err error) error {
	if exc, ok := err.(*Exception); ok {
		return exc.Reason
	}
	return err
}

// OK is a pointer to the zero value of Exception, representing the absence of
// exception.
var OK = &Exception{}

// Error returns the message of the cause of the exception.
func (exc *Exception) Error() string {
	return exc.Reason.Error()
}

// Show shows the exception.
func (exc *Exception) Show(indent string) string {
	buf := new(bytes.Buffer)

	var causeDescription string
	if shower, ok := exc.Reason.(diag.Shower); ok {
		causeDescription = shower.Show(indent)
	} else {
		causeDescription = "\033[31;1m" + exc.Reason.Error() + "\033[m"
	}
	fmt.Fprintf(buf, "Exception: %s\n", causeDescription)

	if exc.StackTrace.next == nil {
		buf.WriteString(exc.StackTrace.head.ShowCompact(indent))
	} else {
		buf.WriteString(indent + "Traceback:")
		for tb := exc.StackTrace; tb != nil; tb = tb.next {
			buf.WriteString("\n" + indent + "  ")
			buf.WriteString(tb.head.Show(indent + "    "))
		}
	}

	if pipeExcs, ok := exc.Reason.(PipelineError); ok {
		buf.WriteString("\n" + indent + "Caused by:")
		for _, e := range pipeExcs.Errors {
			if e == OK {
				continue
			}
			buf.WriteString("\n" + indent + "  " + e.Show(indent+"  "))
		}
	}

	return buf.String()
}

// Kind returns "exception".
func (exc *Exception) Kind() string {
	return "exception"
}

// Repr returns a representation of the exception. It is lossy in that it does
// not preserve the stacktrace.
func (exc *Exception) Repr(indent int) string {
	if exc.Reason == nil {
		return "$ok"
	}
	if r, ok := exc.Reason.(vals.Reprer); ok {
		return r.Repr(indent)
	}
	return "?(fail " + parse.Quote(exc.Reason.Error()) + ")"
}

// Equal compares by address.
func (exc *Exception) Equal(rhs interface{}) bool {
	return exc == rhs
}

// Hash returns the hash of the address.
func (exc *Exception) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(exc))
}

// Bool returns whether this exception has a nil cause; that is, it is $ok.
func (exc *Exception) Bool() bool {
	return exc.Reason == nil
}

func (exc *Exception) Fields() vals.StructMap { return excFields{exc} }

type excFields struct{ e *Exception }

func (excFields) IsStructMap()    {}
func (f excFields) Reason() error { return f.e.Reason }

// PipelineError represents the errors of pipelines, in which multiple commands
// may error.
type PipelineError struct {
	Errors []*Exception
}

// Repr returns a representation of the pipeline error, using the multi-error builtin.
func (pe PipelineError) Repr(indent int) string {
	// TODO Make a more generalized ListReprBuilder and use it here.
	b := new(bytes.Buffer)
	b.WriteString("?(multi-error")
	elemIndent := indent + len("?(multi-error ")
	for _, e := range pe.Errors {
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

// Error returns a plain text representation of the pipeline error.
func (pe PipelineError) Error() string {
	b := new(bytes.Buffer)
	b.WriteString("(")
	for i, e := range pe.Errors {
		if i > 0 {
			b.WriteString(" | ")
		}
		if e == nil || e.Reason == nil {
			b.WriteString("<nil>")
		} else {
			b.WriteString(e.Error())
		}
	}
	b.WriteString(")")
	return b.String()
}

// Builds a suitable error from pipeline results:
//
// * if all elements are either nil or OK, return nil.
//
// * If there is exactly non-nil non-OK Exception, return it.
//
// * Otherwise return a PipelineError built from the slice, with nil items
//   turned into OK's for easier access from elvishscript.
func makePipelineError(excs []*Exception) error {
	newexcs := make([]*Exception, len(excs))
	notOK, lastNotOK := 0, 0
	for i, e := range excs {
		if e == nil {
			newexcs[i] = OK
		} else {
			newexcs[i] = e
			if e.Reason != nil {
				notOK++
				lastNotOK = i
			}
		}
	}
	switch notOK {
	case 0:
		return nil
	case 1:
		return newexcs[lastNotOK]
	default:
		return PipelineError{newexcs}
	}
}

// Flow is a special type of error used for control flows.
type Flow uint

// Control flows.
const (
	Return Flow = iota
	Break
	Continue
)

var flowNames = [...]string{
	"return", "break", "continue",
}

// Repr returns a representation of the flow "error".
func (f Flow) Repr(int) string {
	return "?(" + f.Error() + ")"
}

func (f Flow) Error() string {
	if f >= Flow(len(flowNames)) {
		return fmt.Sprintf("!(BAD FLOW: %d)", f)
	}
	return flowNames[f]
}

// Show shows the flow "error".
func (f Flow) Show(string) string {
	return "\033[33;1m" + f.Error() + "\033[m"
}

// ExternalCmdExit contains the exit status of external commands.
type ExternalCmdExit struct {
	syscall.WaitStatus
	CmdName string
	Pid     int
}

// NewExternalCmdExit constructs an error for representing a non-zero exit from
// an external command.
func NewExternalCmdExit(name string, ws syscall.WaitStatus, pid int) error {
	if ws.Exited() && ws.ExitStatus() == 0 {
		return nil
	}
	return ExternalCmdExit{ws, name, pid}
}

func (exit ExternalCmdExit) Error() string {
	ws := exit.WaitStatus
	quotedName := parse.Quote(exit.CmdName)
	switch {
	case ws.Exited():
		return quotedName + " exited with " + strconv.Itoa(ws.ExitStatus())
	case ws.Signaled():
		causeDescription := quotedName + " killed by signal " + ws.Signal().String()
		if ws.CoreDump() {
			causeDescription += " (core dumped)"
		}
		return causeDescription
	case ws.Stopped():
		causeDescription := quotedName + " stopped by signal " + fmt.Sprintf("%s (pid=%d)", ws.StopSignal(), exit.Pid)
		trap := ws.TrapCause()
		if trap != -1 {
			causeDescription += fmt.Sprintf(" (trapped %v)", trap)
		}
		return causeDescription
	default:
		return fmt.Sprint(quotedName, " has unknown WaitStatus ", ws)
	}
}

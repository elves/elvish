package eval

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hash"
)

// Exception represents an elvish exception. It is both a Value accessible to
// elvishscript, and the type of error returned by public facing evaluation
// methods like (*Evaler)PEval.
type Exception struct {
	Cause     error
	Traceback *stackTrace
}

// OK is a pointer to the zero value of Exception, representing the absence of
// exception.
var OK = &Exception{}

func (exc *Exception) Error() string {
	return exc.Cause.Error()
}

func (exc *Exception) Pprint(indent string) string {
	buf := new(bytes.Buffer)

	var causeDescription string
	if pprinter, ok := exc.Cause.(util.Pprinter); ok {
		causeDescription = pprinter.Pprint(indent)
	} else {
		causeDescription = "\033[31;1m" + exc.Cause.Error() + "\033[m"
	}
	fmt.Fprintf(buf, "Exception: %s\n", causeDescription)

	if exc.Traceback.next == nil {
		buf.WriteString(exc.Traceback.entry.PprintCompact(indent))
	} else {
		buf.WriteString(indent + "Traceback:")
		for tb := exc.Traceback; tb != nil; tb = tb.next {
			buf.WriteString("\n" + indent + "  ")
			buf.WriteString(tb.entry.Pprint(indent + "    "))
		}
	}

	if pipeExcs, ok := exc.Cause.(PipelineError); ok {
		buf.WriteString("\n" + indent + "Caused by:")
		for _, e := range pipeExcs.Errors {
			if e == OK {
				continue
			}
			buf.WriteString("\n" + indent + "  " + e.Pprint(indent+"  "))
		}
	}

	return buf.String()
}

func (exc *Exception) Kind() string {
	return "exception"
}

func (exc *Exception) Repr(indent int) string {
	if exc.Cause == nil {
		return "$ok"
	}
	if r, ok := exc.Cause.(vals.Reprer); ok {
		return r.Repr(indent)
	}
	return "?(fail " + parse.Quote(exc.Cause.Error()) + ")"
}

// Equal compares by identity.
func (exc *Exception) Equal(rhs interface{}) bool {
	return exc == rhs
}

func (exc *Exception) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(exc))
}

func (exc *Exception) Bool() bool {
	return exc.Cause == nil
}

func (exc *Exception) Index(k interface{}) (interface{}, bool) {
	// TODO: Access to Traceback
	switch k {
	case "cause":
		return exc.Cause, true
	default:
		return nil, false
	}
}

func (exc *Exception) IterateKeys(f func(interface{}) bool) {
	vals.Feed(f, "cause")
}

// PipelineError represents the errors of pipelines, in which multiple commands
// may error.
type PipelineError struct {
	Errors []*Exception
}

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

func (pe PipelineError) Error() string {
	b := new(bytes.Buffer)
	b.WriteString("(")
	for i, e := range pe.Errors {
		if i > 0 {
			b.WriteString(" | ")
		}
		if e == nil || e.Cause == nil {
			b.WriteString("<nil>")
		} else {
			b.WriteString(e.Error())
		}
	}
	b.WriteString(")")
	return b.String()
}

// ComposeExceptionsFromPipeline takes a slice of Exception pointers and
// composes a suitable error. If all elements of the slice are either nil or OK,
// a nil is returned. If there is exactly non-nil non-OK Exception, it is
// returned. Otherwise, a PipelineError built from the slice is returned, with
// nil items turned into OK's for easier access from elvishscript.
func ComposeExceptionsFromPipeline(excs []*Exception) error {
	newexcs := make([]*Exception, len(excs))
	notOK, lastNotOK := 0, 0
	for i, e := range excs {
		if e == nil {
			newexcs[i] = OK
		} else {
			newexcs[i] = e
			if e.Cause != nil {
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

func (f Flow) Repr(int) string {
	return "?(" + f.Error() + ")"
}

func (f Flow) Error() string {
	if f >= Flow(len(flowNames)) {
		return fmt.Sprintf("!(BAD FLOW: %v)", f)
	}
	return flowNames[f]
}

func (f Flow) Pprint(string) string {
	return "\033[33;1m" + f.Error() + "\033[m"
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

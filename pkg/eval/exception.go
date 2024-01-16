package eval

import (
	"bytes"
	"fmt"
	"strconv"
	"syscall"
	"unsafe"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/persistent/hash"
)

// Exception represents exceptions. It is both a Value accessible to Elvish
// code, and can be returned by methods like (*Evaler).Eval.
type Exception interface {
	error
	diag.Shower
	Reason() error
	StackTrace() *StackTrace
	// This is not strictly necessary, but it makes sure that there is only one
	// implementation of Exception, so that the compiler may de-virtualize this
	// interface.
	isException()
}

// NewException creates a new Exception.
func NewException(reason error, stackTrace *StackTrace) Exception {
	return &exception{reason, stackTrace}
}

// Implementation of the Exception interface.
type exception struct {
	reason     error
	stackTrace *StackTrace
}

var _ vals.PseudoMap = &exception{}

// StackTrace represents a stack trace as a linked list of diag.Context. The
// head is the innermost stack.
//
// Since pipelines can call multiple functions in parallel, all the StackTrace
// nodes form a DAG.
type StackTrace struct {
	Head *diag.Context
	Next *StackTrace
}

// Reason returns the Reason field if err is an Exception. Otherwise it returns
// err itself.
func Reason(err error) error {
	if exc, ok := err.(*exception); ok {
		return exc.reason
	}
	return err
}

// OK is a pointer to a special value of Exception that represents the absence
// of exception.
var OK = &exception{}

func (exc *exception) isException() {}

func (exc *exception) Reason() error { return exc.reason }

func (exc *exception) StackTrace() *StackTrace { return exc.stackTrace }

// Error returns the message of the cause of the exception.
func (exc *exception) Error() string { return exc.reason.Error() }

var (
	exceptionCauseStartMarker = "\033[31;1m"
	exceptionCauseEndMarker   = "\033[m"
)

// Show shows the exception.
func (exc *exception) Show(indent string) string {
	buf := new(bytes.Buffer)

	var causeDescription string
	if shower, ok := exc.reason.(diag.Shower); ok {
		causeDescription = shower.Show(indent)
	} else if exc.reason == nil {
		causeDescription = "ok"
	} else {
		causeDescription = exceptionCauseStartMarker + exc.reason.Error() + exceptionCauseEndMarker
	}
	fmt.Fprintf(buf, "Exception: %s", causeDescription)

	if exc.stackTrace != nil {
		for tb := exc.stackTrace; tb != nil; tb = tb.Next {
			buf.WriteString("\n" + indent + "  ")
			buf.WriteString(tb.Head.Show(indent + "  "))
		}
	}

	if pipeExcs, ok := exc.reason.(PipelineError); ok {
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
func (exc *exception) Kind() string {
	return "exception"
}

// Repr returns a representation of the exception. It is lossy in that it does
// not preserve the stacktrace.
func (exc *exception) Repr(indent int) string {
	if exc.reason == nil {
		return "$ok"
	}
	return "[^exception &reason=" + vals.Repr(exc.reason, indent+1) + " &stack-trace=<...>]"
}

// Equal compares by address.
func (exc *exception) Equal(rhs any) bool {
	return exc == rhs
}

// Hash returns the hash of the address.
func (exc *exception) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(exc))
}

// Bool returns whether this exception has a nil cause; that is, it is $ok.
func (exc *exception) Bool() bool {
	return exc.reason == nil
}

func (exc *exception) Fields() vals.StructMap { return excFields{exc} }

type excFields struct{ e *exception }

func (excFields) IsStructMap()              {}
func (f excFields) Reason() error           { return f.e.reason }
func (f excFields) StackTrace() *StackTrace { return f.e.stackTrace }

// PipelineError represents the errors of pipelines, in which multiple commands
// may error.
type PipelineError struct {
	Errors []Exception
}

var _ vals.PseudoMap = PipelineError{}

// Error returns a plain text representation of the pipeline error.
func (pe PipelineError) Error() string {
	b := new(bytes.Buffer)
	b.WriteString("(")
	for i, e := range pe.Errors {
		if i > 0 {
			b.WriteString(" | ")
		}
		if e == nil || e.Reason() == nil {
			b.WriteString("<nil>")
		} else {
			b.WriteString(e.Error())
		}
	}
	b.WriteString(")")
	return b.String()
}

// MakePipelineError builds an error from the execution results of multiple
// commands in a pipeline.
//
// If all elements are either nil or OK, it returns nil. If there is exactly
// non-nil non-OK Exception, it returns it. Otherwise, it return a PipelineError
// built from the slice, with nil items turned into OK's for easier access from
// Elvish code.
func MakePipelineError(excs []Exception) error {
	newexcs := make([]Exception, len(excs))
	notOK, lastNotOK := 0, 0
	for i, e := range excs {
		if e == nil {
			newexcs[i] = OK
		} else {
			newexcs[i] = e
			if e.Reason() != nil {
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

func (pe PipelineError) Kind() string           { return "pipeline-error" }
func (pe PipelineError) Fields() vals.StructMap { return peFields{pe} }

type peFields struct{ pe PipelineError }

func (peFields) IsStructMap() {}

func (f peFields) Type() string { return "pipeline" }

func (f peFields) Exceptions() vals.List {
	li := vals.EmptyList
	for _, exc := range f.pe.Errors {
		li = li.Conj(exc)
	}
	return li
}

// Flow is a special type of error used for control flows.
type Flow uint

var _ vals.PseudoMap = Flow(0)

// Control flows.
const (
	Return Flow = iota
	Break
	Continue
)

var flowNames = [...]string{
	"return", "break", "continue",
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

func (f Flow) Kind() string           { return "flow-error" }
func (f Flow) Fields() vals.StructMap { return flowFields{f} }

type flowFields struct{ f Flow }

func (flowFields) IsStructMap() {}

func (f flowFields) Type() string { return "flow" }
func (f flowFields) Name() string { return f.f.Error() }

// ExternalCmdExit contains the exit status of external commands.
type ExternalCmdExit struct {
	syscall.WaitStatus
	CmdName string
	Pid     int
}

var _ vals.PseudoMap = ExternalCmdExit{}

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

func (exit ExternalCmdExit) Kind() string {
	return "external-cmd-error"
}

func (exit ExternalCmdExit) Fields() vals.StructMap {
	ws := exit.WaitStatus
	f := exitFieldsCommon{exit}
	switch {
	case ws.Exited():
		return exitFieldsExited{f}
	case ws.Signaled():
		return exitFieldsSignaled{f}
	case ws.Stopped():
		return exitFieldsStopped{f}
	default:
		return exitFieldsUnknown{f}
	}
}

type exitFieldsCommon struct{ e ExternalCmdExit }

func (exitFieldsCommon) IsStructMap()      {}
func (f exitFieldsCommon) CmdName() string { return f.e.CmdName }
func (f exitFieldsCommon) Pid() string     { return strconv.Itoa(f.e.Pid) }

type exitFieldsExited struct{ exitFieldsCommon }

func (exitFieldsExited) Type() string         { return "external-cmd/exited" }
func (f exitFieldsExited) ExitStatus() string { return strconv.Itoa(f.e.ExitStatus()) }

type exitFieldsSignaled struct{ exitFieldsCommon }

func (f exitFieldsSignaled) Type() string         { return "external-cmd/signaled" }
func (f exitFieldsSignaled) SignalName() string   { return f.e.Signal().String() }
func (f exitFieldsSignaled) SignalNumber() string { return strconv.Itoa(int(f.e.Signal())) }
func (f exitFieldsSignaled) CoreDumped() bool     { return f.e.CoreDump() }

type exitFieldsStopped struct{ exitFieldsCommon }

func (f exitFieldsStopped) Type() string         { return "external-cmd/stopped" }
func (f exitFieldsStopped) SignalName() string   { return f.e.StopSignal().String() }
func (f exitFieldsStopped) SignalNumber() string { return strconv.Itoa(int(f.e.StopSignal())) }
func (f exitFieldsStopped) TrapCause() int       { return f.e.TrapCause() }

type exitFieldsUnknown struct{ exitFieldsCommon }

func (exitFieldsUnknown) Type() string { return "external-cmd/unknown" }

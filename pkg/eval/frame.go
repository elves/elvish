package eval

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/prog"
	"github.com/elves/elvish/pkg/strutil"
)

// Frame contains information of the current running function, akin to a call
// frame in native CPU execution. A Frame is only modified during and very
// shortly after creation; new Frame's are "forked" when needed.
type Frame struct {
	*Evaler

	srcMeta parse.Source

	local, up Ns

	intCh <-chan struct{}
	ports []*Port

	traceback *StackTrace

	background bool
}

// NewTopFrame creates a top-level Frame.
//
// TODO(xiaq): This should be a method on the Evaler.
func NewTopFrame(ev *Evaler, src parse.Source, ports []*Port) *Frame {
	return &Frame{
		ev, src,
		ev.Global, make(Ns),
		nil, ports,
		nil, false,
	}
}

// SetLocal changes the local scope of the Frame.
func (fm *Frame) SetLocal(ns Ns) {
	fm.local = ns
}

// Close releases resources allocated for this frame. It always returns a nil
// error. It may be called only once.
func (fm *Frame) Close() error {
	for _, port := range fm.ports {
		port.Close()
	}
	return nil
}

// InputChan returns a channel from which input can be read.
func (fm *Frame) InputChan() chan interface{} {
	return fm.ports[0].Chan
}

// InputFile returns a file from which input can be read.
func (fm *Frame) InputFile() *os.File {
	return fm.ports[0].File
}

// OutputChan returns a channel onto which output can be written.
func (fm *Frame) OutputChan() chan<- interface{} {
	return fm.ports[1].Chan
}

// OutputFile returns a file onto which output can be written.
func (fm *Frame) OutputFile() *os.File {
	return fm.ports[1].File
}

// ErrorFile returns a file onto which error messages can be written.
func (fm *Frame) ErrorFile() *os.File {
	return fm.ports[2].File
}

// IterateInputs calls the passed function for each input element.
func (fm *Frame) IterateInputs(f func(interface{})) {
	var w sync.WaitGroup
	inputs := make(chan interface{})

	w.Add(2)
	go func() {
		linesToChan(fm.InputFile(), inputs)
		w.Done()
	}()
	go func() {
		for v := range fm.ports[0].Chan {
			inputs <- v
		}
		w.Done()
	}()
	go func() {
		w.Wait()
		close(inputs)
	}()

	for v := range inputs {
		f(v)
	}
}

func linesToChan(r io.Reader, ch chan<- interface{}) {
	filein := bufio.NewReader(r)
	for {
		line, err := filein.ReadString('\n')
		if line != "" {
			ch <- strutil.ChopLineEnding(line)
		}
		if err != nil {
			if err != io.EOF {
				logger.Println("error on reading:", err)
			}
			break
		}
	}
}

// fork returns a modified copy of ec. The ports are forked, and the name is
// changed to the given value. Other fields are copied shallowly.
func (fm *Frame) fork(name string) *Frame {
	newPorts := make([]*Port, len(fm.ports))
	for i, p := range fm.ports {
		if p != nil {
			newPorts[i] = p.Fork()
		}
	}
	return &Frame{
		fm.Evaler, fm.srcMeta,
		fm.local, fm.up,
		fm.intCh, newPorts,
		fm.traceback, fm.background,
	}
}

// Eval evaluates an Op. It is like eval except that it sets fm.srcMeta
// temporarily to op.src during the evaluation.
func (fm *Frame) Eval(op Op) error {
	oldSrc := fm.srcMeta
	fm.srcMeta = op.Src
	defer func() {
		fm.srcMeta = oldSrc
	}()
	return op.Inner.exec(fm)
}

// CaptureOutput captures the output of a given callback that operates on a Frame.
func (fm *Frame) CaptureOutput(f func(*Frame) error) ([]interface{}, error) {
	return captureOutput(fm, f)
}

// PipeOutput calls a callback with output piped to the given output handlers.
func (fm *Frame) PipeOutput(f func(*Frame) error, valuesCb func(<-chan interface{}), bytesCb func(*os.File)) error {
	return pipeOutput(fm, f, valuesCb, bytesCb)
}

func (fm *Frame) addTraceback(r diag.Ranger) *StackTrace {
	return &StackTrace{
		Head: diag.NewContext(fm.srcMeta.Name, fm.srcMeta.Code, r.Range()),
		Next: fm.traceback,
	}
}

// Returns an Exception with specified range and cause.
func (fm *Frame) errorp(r diag.Ranger, e error) error {
	switch e := e.(type) {
	case nil:
		return nil
	case *Exception:
		return e
	default:
		return &Exception{e, &StackTrace{
			Head: diag.NewContext(fm.srcMeta.Name, fm.srcMeta.Code, r.Range()),
			Next: fm.traceback,
		}}
	}
}

// Returns an Exception with specified range and error text.
func (fm *Frame) errorpf(r diag.Ranger, format string, args ...interface{}) error {
	return fm.errorp(r, fmt.Errorf(format, args...))
}

// Deprecate shows a deprecation message. The message is not shown if the same
// deprecation message has been shown for the same location before.
func (fm *Frame) Deprecate(msg string, ctx *diag.Context) {
	if ctx == nil {
		ctx = fm.traceback.Head
	}
	dep := deprecation{ctx.Name, ctx.Ranging, msg}
	if prog.ShowDeprecations && fm.deprecations.register(dep) {
		err := diag.Error{
			Type: "deprecation", Message: dep.message, Context: *ctx}
		fm.ports[2].File.WriteString(err.Show("") + "\n")
	}
}

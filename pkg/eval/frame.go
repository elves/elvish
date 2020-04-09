package eval

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/pkg/diag"
)

// Frame contains information of the current running function, aknin to a call
// frame in native CPU execution. A Frame is only modified during and very
// shortly after creation; new Frame's are "forked" when needed.
type Frame struct {
	*Evaler

	srcMeta *Source

	local, up Ns

	intCh chan struct{}
	ports []*Port

	traceback *stackTrace

	background bool
}

// NewTopFrame creates a top-level Frame.
//
// TODO(xiaq): This should be a method on the Evaler.
func NewTopFrame(ev *Evaler, src *Source, ports []*Port) *Frame {
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

// IterateInputs calls the passed function for each input element.
func (fm *Frame) IterateInputs(f func(interface{})) {
	var w sync.WaitGroup
	inputs := make(chan interface{})

	w.Add(2)
	go func() {
		linesToChan(fm.ports[0].File, inputs)
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
			ch <- strings.TrimSuffix(line, "\n")
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

// CaptureOutput calls a function with the given arguments and options,
// capturing and returning the output. It does so in a protected environment so
// that exceptions thrown are wrapped in an Error.
func (fm *Frame) CaptureOutput(fn Callable, args []interface{}, opts map[string]interface{}) (vs []interface{}, err error) {
	// XXX There is no source.
	opFunc := func(f *Frame) error {
		return fn.Call(f, args, opts)
	}
	return pcaptureOutput(fm, effectOp{funcOp(opFunc), diag.Ranging{From: -1, To: -1}})
}

// CallWithOutputCallback calls a function with the given arguments and options,
// feeding the outputs to the given callbacks. It does so in a protected
// environment so that exceptions thrown are wrapped in an Error.
func (fm *Frame) CallWithOutputCallback(fn Callable, args []interface{}, opts map[string]interface{}, valuesCb func(<-chan interface{}), bytesCb func(*os.File)) error {
	// XXX There is no source.
	opFunc := func(f *Frame) error {
		return fn.Call(f, args, opts)
	}
	return pcaptureOutputInner(fm, effectOp{funcOp(opFunc), diag.Ranging{From: -1, To: -1}}, valuesCb, bytesCb)
}

// ExecWithOutputCallback executes an Op, feeding the outputs to the given
// callbacks.
func (fm *Frame) ExecWithOutputCallback(op Op, valuesCb func(<-chan interface{}), bytesCb func(*os.File)) error {
	return pcaptureOutputInner(fm, op.Inner, valuesCb, bytesCb)
}

func (fm *Frame) addTraceback(r diag.Ranger) *stackTrace {
	return &stackTrace{
		head: diag.NewContext(fm.srcMeta.Name, fm.srcMeta.Code, r.Range()),
		next: fm.traceback,
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
		return &Exception{e, &stackTrace{
			head: diag.NewContext(fm.srcMeta.Name, fm.srcMeta.Code, r.Range()),
			next: fm.traceback,
		}}
	}
}

// Returns an Exception with specified range and error text.
func (fm *Frame) errorpf(r diag.Ranger, format string, args ...interface{}) error {
	return fm.errorp(r, fmt.Errorf(format, args...))
}

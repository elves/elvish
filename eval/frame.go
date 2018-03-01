package eval

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/util"
)

// Frame contains information of the current running function, aknin to a call
// frame in native CPU execution. A Frame is only modified during and very
// shortly after creation; new Frame's are "forked" when needed.
type Frame struct {
	*Evaler
	srcMeta *Source

	local, up Ns
	ports     []*Port

	begin, end int
	traceback  *stackTrace

	background bool
}

// NewTopFrame creates a top-level Frame.
func NewTopFrame(ev *Evaler, src *Source, ports []*Port) *Frame {
	return &Frame{
		ev, src,
		ev.Global, make(Ns),
		ports,
		0, len(src.code), nil, false,
	}
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
		newPorts[i] = p.Fork()
	}
	return &Frame{
		fm.Evaler, fm.srcMeta,
		fm.local, fm.up,
		newPorts,
		fm.begin, fm.end, fm.traceback, fm.background,
	}
}

// Eval evaluates an op. It does so in a protected environment so that
// exceptions thrown are wrapped in an Error.
func (fm *Frame) Eval(op Op) (err error) {
	defer catch(&err, fm)
	e := op.Exec(fm)
	if e != nil {
		if exc, ok := e.(*Exception); ok {
			return exc
		}
		return fm.makeException(e)
	}
	return nil
}

// Call calls a function with the given arguments and options. It does so in a
// protected environment so that exceptions thrown are wrapped in an Error.
func (fm *Frame) Call(f Callable, args []interface{}, opts map[string]interface{}) (err error) {
	defer catch(&err, fm)
	e := f.Call(fm, args, opts)
	if e != nil {
		if exc, ok := e.(*Exception); ok {
			return exc
		}
		return fm.makeException(e)
	}
	return nil
}

// CaptureOutput calls a function with the given arguments and options,
// capturing and returning the output. It does so in a protected environment so
// that exceptions thrown are wrapped in an Error.
func (fm *Frame) CaptureOutput(fn Callable, args []interface{}, opts map[string]interface{}) (vs []interface{}, err error) {
	// XXX There is no source.
	opFunc := func(f *Frame) error {
		return fn.Call(f, args, opts)
	}
	return pcaptureOutput(fm, Op{funcOp(opFunc), -1, -1})
}

// CallWithOutputCallback calls a function with the given arguments and options,
// feeding the outputs to the given callbacks. It does so in a protected
// environment so that exceptions thrown are wrapped in an Error.
func (fm *Frame) CallWithOutputCallback(fn Callable, args []interface{}, opts map[string]interface{}, valuesCb func(<-chan interface{}), bytesCb func(*os.File)) error {
	// XXX There is no source.
	opFunc := func(f *Frame) error {
		return fn.Call(f, args, opts)
	}
	return pcaptureOutputInner(fm, Op{funcOp(opFunc), -1, -1}, valuesCb, bytesCb)
}

func catch(perr *error, fm *Frame) {
	// NOTE: We have to duplicate instead of calling util.Catch here, since
	// recover can only catch a panic when called directly from a deferred
	// function.
	r := recover()
	if r == nil {
		return
	}
	if exc, ok := r.(util.Thrown); ok {
		err := exc.Wrapped
		if _, ok := err.(*Exception); !ok {
			err = fm.makeException(err)
		}
		*perr = err
	} else if r != nil {
		panic(r)
	}
}

// makeException turns an error into an Exception by adding traceback.
func (fm *Frame) makeException(e error) *Exception {
	return &Exception{e, fm.addTraceback()}
}

func (fm *Frame) addTraceback() *stackTrace {
	return &stackTrace{
		entry: &util.SourceRange{
			Name: fm.srcMeta.describePath(), Source: fm.srcMeta.code,
			Begin: fm.begin, End: fm.end,
		},
		next: fm.traceback,
	}
}

// errorpf stops the ec.eval immediately by panicking with a diagnostic message.
// The panic is supposed to be caught by ec.eval.
func (fm *Frame) errorpf(begin, end int, format string, args ...interface{}) {
	fm.begin, fm.end = begin, end
	throwf(format, args...)
}

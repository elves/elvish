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

// InputChan returns a channel from which input can be read.
func (ec *Frame) InputChan() chan interface{} {
	return ec.ports[0].Chan
}

// InputFile returns a file from which input can be read.
func (ec *Frame) InputFile() *os.File {
	return ec.ports[0].File
}

// OutputChan returns a channel onto which output can be written.
func (ec *Frame) OutputChan() chan<- interface{} {
	return ec.ports[1].Chan
}

// OutputFile returns a file onto which output can be written.
func (ec *Frame) OutputFile() *os.File {
	return ec.ports[1].File
}

// IterateInputs calls the passed function for each input element.
func (ec *Frame) IterateInputs(f func(interface{})) {
	var w sync.WaitGroup
	inputs := make(chan interface{})

	w.Add(2)
	go func() {
		linesToChan(ec.ports[0].File, inputs)
		w.Done()
	}()
	go func() {
		for v := range ec.ports[0].Chan {
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
func (ec *Frame) fork(name string) *Frame {
	newPorts := make([]*Port, len(ec.ports))
	for i, p := range ec.ports {
		newPorts[i] = p.Fork()
	}
	return &Frame{
		ec.Evaler, ec.srcMeta,
		ec.local, ec.up,
		newPorts,
		ec.begin, ec.end, ec.traceback, ec.background,
	}
}

// PEval evaluates an op in a protected environment so that calls to errorf are
// wrapped in an Error.
func (ec *Frame) PEval(op Op) (err error) {
	defer catch(&err, ec)
	e := op.Exec(ec)
	if e != nil {
		if exc, ok := e.(*Exception); ok {
			return exc
		}
		return ec.makeException(e)
	}
	return nil
}

func (ec *Frame) PCall(f Callable, args []interface{}, opts map[string]interface{}) (err error) {
	defer catch(&err, ec)
	e := f.Call(ec, args, opts)
	if e != nil {
		if exc, ok := e.(*Exception); ok {
			return exc
		}
		return ec.makeException(e)
	}
	return nil
}

func (ec *Frame) PCaptureOutput(fn Callable, args []interface{}, opts map[string]interface{}) (vs []interface{}, err error) {
	// XXX There is no source.
	opFunc := func(f *Frame) error {
		return fn.Call(f, args, opts)
	}
	return pcaptureOutput(ec, Op{funcOp(opFunc), -1, -1})
}

func (ec *Frame) PCaptureOutputInner(fn Callable, args []interface{}, opts map[string]interface{}, valuesCb func(<-chan interface{}), bytesCb func(*os.File)) error {
	// XXX There is no source.
	opFunc := func(f *Frame) error {
		return fn.Call(f, args, opts)
	}
	return pcaptureOutputInner(ec, Op{funcOp(opFunc), -1, -1}, valuesCb, bytesCb)
}

func catch(perr *error, ec *Frame) {
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
			err = ec.makeException(err)
		}
		*perr = err
	} else if r != nil {
		panic(r)
	}
}

// makeException turns an error into an Exception by adding traceback.
func (ec *Frame) makeException(e error) *Exception {
	return &Exception{e, ec.addTraceback()}
}

func (ec *Frame) addTraceback() *stackTrace {
	return &stackTrace{
		entry: &util.SourceRange{
			Name: ec.srcMeta.describePath(), Source: ec.srcMeta.code,
			Begin: ec.begin, End: ec.end,
		},
		next: ec.traceback,
	}
}

// errorpf stops the ec.eval immediately by panicking with a diagnostic message.
// The panic is supposed to be caught by ec.eval.
func (ec *Frame) errorpf(begin, end int, format string, args ...interface{}) {
	ec.begin, ec.end = begin, end
	throwf(format, args...)
}

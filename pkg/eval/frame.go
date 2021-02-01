package eval

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/strutil"
)

// Frame contains information of the current running function, akin to a call
// frame in native CPU execution. A Frame is only modified during and very
// shortly after creation; new Frame's are "forked" when needed.
type Frame struct {
	Evaler *Evaler

	srcMeta parse.Source

	local, up *Ns

	intCh <-chan struct{}
	ports []*Port

	traceback *StackTrace

	background bool
}

// PrepareEval prepares a piece of code for evaluation in a copy of the current
// Frame. If r is not nil, it is added to the traceback of the evaluation
// context. If ns is not nil, it is used in place of the current local namespace
// as the namespace to evaluate the code in.
//
// If there is any parse error or compilation error, it returns a nil *Ns, nil
// function and the error. If there is no parse error or compilation error, it
// returns the altered local namespace, function that can be called to actuate
// the evaluation, and a nil error.
func (fm *Frame) PrepareEval(src parse.Source, r diag.Ranger, ns *Ns) (*Ns, func() Exception, error) {
	tree, err := parse.Parse(src, parse.Config{WarningWriter: fm.ErrorFile()})
	if err != nil {
		return nil, nil, err
	}
	local := fm.local
	if ns != nil {
		local = ns
	}
	traceback := fm.traceback
	if r != nil {
		traceback = fm.addTraceback(r)
	}
	newFm := &Frame{
		fm.Evaler, src, local, new(Ns), fm.intCh, fm.ports, traceback, fm.background}
	op, err := compile(newFm.Evaler.Builtin().static(), local.static(), tree, fm.ErrorFile())
	if err != nil {
		return nil, nil, err
	}
	newLocal, exec := op.prepare(newFm)
	return newLocal, exec, nil
}

// Eval evaluates a piece of code in a copy of the current Frame. It returns the
// altered local namespace, and any parse error, compilation error or exception.
//
// See PrepareEval for a description of the arguments.
func (fm *Frame) Eval(src parse.Source, r diag.Ranger, ns *Ns) (*Ns, error) {
	newLocal, exec, err := fm.PrepareEval(src, r, ns)
	if err != nil {
		return nil, err
	}
	return newLocal, exec()
}

// Close releases resources allocated for this frame. It always returns a nil
// error. It may be called only once.
func (fm *Frame) Close() error {
	for _, port := range fm.ports {
		port.close()
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
			newPorts[i] = p.fork()
		}
	}
	return &Frame{
		fm.Evaler, fm.srcMeta,
		fm.local, fm.up,
		fm.intCh, newPorts,
		fm.traceback, fm.background,
	}
}

// A shorthand for forking a frame and setting the output port.
func (fm *Frame) forkWithOutput(name string, p *Port) *Frame {
	newFm := fm.fork(name)
	newFm.ports[1] = p
	return newFm
}

// CaptureOutput captures the output of a given callback that operates on a Frame.
func (fm *Frame) CaptureOutput(f func(*Frame) error) ([]interface{}, error) {
	outPort, collect, err := CapturePort()
	if err != nil {
		return nil, err
	}
	err = f(fm.forkWithOutput("[output capture]", outPort))
	return collect(), err
}

// PipeOutput calls a callback with output piped to the given output handlers.
func (fm *Frame) PipeOutput(f func(*Frame) error, vCb func(<-chan interface{}), bCb func(*os.File)) error {
	outPort, done, err := PipePort(vCb, bCb)
	if err != nil {
		return err
	}
	err = f(fm.forkWithOutput("[output pipe]", outPort))
	done()
	return err
}

func (fm *Frame) addTraceback(r diag.Ranger) *StackTrace {
	return &StackTrace{
		Head: diag.NewContext(fm.srcMeta.Name, fm.srcMeta.Code, r.Range()),
		Next: fm.traceback,
	}
}

// Returns an Exception with specified range and cause.
func (fm *Frame) errorp(r diag.Ranger, e error) Exception {
	switch e := e.(type) {
	case nil:
		return nil
	case Exception:
		return e
	default:
		return &exception{e, &StackTrace{
			Head: diag.NewContext(fm.srcMeta.Name, fm.srcMeta.Code, r.Range()),
			Next: fm.traceback,
		}}
	}
}

// Returns an Exception with specified range and error text.
func (fm *Frame) errorpf(r diag.Ranger, format string, args ...interface{}) Exception {
	return fm.errorp(r, fmt.Errorf(format, args...))
}

// Deprecate shows a deprecation message. The message is not shown if the same
// deprecation message has been shown for the same location before.
func (fm *Frame) Deprecate(msg string, ctx *diag.Context, minLevel int) {
	if prog.DeprecationLevel < minLevel {
		return
	}
	if ctx == nil {
		fmt.Fprintf(fm.ErrorFile(), "deprecation: \033[31;1m%s\033[m\n", msg)
		return
	}
	if fm.Evaler.registerDeprecation(deprecation{ctx.Name, ctx.Ranging, msg}) {
		err := diag.Error{Type: "deprecation", Message: msg, Context: *ctx}
		fm.ErrorFile().WriteString(err.Show("") + "\n")
	}
}

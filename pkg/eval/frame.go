package eval

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/strutil"
)

// Frame contains information of the current running function, akin to a call
// frame in native CPU execution. A Frame is only modified during and very
// shortly after creation; new Frame's are "forked" when needed.
type Frame struct {
	Evaler *Evaler

	src parse.Source

	local, up *Ns
	defers    *[]func(*Frame) Exception

	// The godoc of the context package states:
	//
	// > Do not store Contexts inside a struct type; instead, pass a Context
	// > explicitly to each function that needs it.
	//
	// However, that advice is considered by many to be overly aggressive
	// (https://github.com/golang/go/issues/22602). The Frame struct doesn't fit
	// the "parameter struct" definition in that discussion, but it is itself is
	// a "context struct". Storing a Context inside it seems fine.
	ctx   context.Context
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
		fm.Evaler, src, local, new(Ns), nil, fm.ctx, fm.ports, traceback, fm.background}
	op, _, err := compile(fm.Evaler.Builtin().static(), local.static(), nil, tree, fm.ErrorFile())
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
func (fm *Frame) InputChan() chan any {
	return fm.ports[0].Chan
}

// InputFile returns a file from which input can be read.
func (fm *Frame) InputFile() *os.File {
	return fm.ports[0].File
}

// ValueOutput returns a handle for writing value outputs.
func (fm *Frame) ValueOutput() ValueOutput {
	p := fm.ports[1]
	return valueOutput{p.Chan, p.sendStop, p.sendError}
}

// ByteOutput returns a handle for writing byte outputs.
func (fm *Frame) ByteOutput() ByteOutput {
	return byteOutput{fm.ports[1].File}
}

// ErrorFile returns a file onto which error messages can be written.
func (fm *Frame) ErrorFile() *os.File {
	return fm.ports[2].File
}

// Port returns port i. If the port doesn't exist, it returns nil
//
// This is a low-level construct that shouldn't be used for writing output; for
// that purpose, use [(*Frame).ValueOutput] and [(*Frame).ByteOutput] instead.
func (fm *Frame) Port(i int) *Port {
	if i >= len(fm.ports) {
		return nil
	}
	return fm.ports[i]
}

// IterateInputs calls the passed function for each input element.
func (fm *Frame) IterateInputs(f func(any)) {
	var wg sync.WaitGroup
	inputs := make(chan any)

	wg.Add(2)
	go func() {
		linesToChan(fm.InputFile(), inputs)
		wg.Done()
	}()
	go func() {
		for v := range fm.ports[0].Chan {
			inputs <- v
		}
		wg.Done()
	}()
	go func() {
		wg.Wait()
		close(inputs)
	}()

	for v := range inputs {
		f(v)
	}
}

func linesToChan(r io.Reader, ch chan<- any) {
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

// Context returns a Context associated with the Frame.
func (fm *Frame) Context() context.Context {
	return fm.ctx
}

// Canceled reports whether the Context of the Frame has been canceled.
func (fm *Frame) Canceled() bool {
	select {
	case <-fm.ctx.Done():
		return true
	default:
		return false
	}
}

// Fork returns a modified copy of fm. The ports are forked, and the name is
// changed to the given value. Other fields are copied shallowly.
func (fm *Frame) Fork() *Frame {
	newPorts := make([]*Port, len(fm.ports))
	for i, p := range fm.ports {
		if p != nil {
			newPorts[i] = p.fork()
		}
	}
	return &Frame{
		fm.Evaler, fm.src,
		fm.local, fm.up, fm.defers,
		fm.ctx, newPorts,
		fm.traceback, fm.background,
	}
}

// A shorthand for forking a frame and setting the output port.
func (fm *Frame) forkWithOutput(p *Port) *Frame {
	newFm := fm.Fork()
	newFm.ports[1] = p
	return newFm
}

// CaptureOutput captures the output of a given callback that operates on a Frame.
func (fm *Frame) CaptureOutput(f func(*Frame) error) ([]any, error) {
	outPort, collect, err := ValueCapturePort()
	if err != nil {
		return nil, err
	}
	err = f(fm.forkWithOutput(outPort))
	return collect(), err
}

// PipeOutput calls a callback with output piped to the given output handlers.
func (fm *Frame) PipeOutput(f func(*Frame) error, vCb func(<-chan any), bCb func(*os.File)) error {
	outPort, done, err := PipePort(vCb, bCb)
	if err != nil {
		return err
	}
	err = f(fm.forkWithOutput(outPort))
	done()
	return err
}

func (fm *Frame) addTraceback(r diag.Ranger) *StackTrace {
	return &StackTrace{
		Head: diag.NewContext(fm.src.Name, fm.src.Code, r.Range()),
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
		if _, ok := e.(errs.SetReadOnlyVar); ok {
			r := r.Range()
			e = errs.SetReadOnlyVar{VarName: fm.src.Code[r.From:r.To]}
		}
		ctx := diag.NewContext(fm.src.Name, fm.src.Code, r)
		return &exception{e, &StackTrace{Head: ctx, Next: fm.traceback}}
	}
}

// Returns an Exception with specified range and error text.
func (fm *Frame) errorpf(r diag.Ranger, format string, args ...any) Exception {
	return fm.errorp(r, fmt.Errorf(format, args...))
}

// Deprecate shows a deprecation message. The message is not shown if the same
// deprecation message has been shown for the same location before.
func (fm *Frame) Deprecate(msg string, ctx *diag.Context, minLevel int) {
	if prog.DeprecationLevel < minLevel {
		return
	}
	if ctx == nil {
		ctx = fm.traceback.Head
	}
	if fm.Evaler.registerDeprecation(deprecation{ctx.Name, ctx.Ranging, msg}) {
		err := diag.Error[deprecationTag]{Message: msg, Context: *ctx}
		fm.ErrorFile().WriteString(err.Show("") + "\n")
	}
}

func (fm *Frame) addDefer(f func(*Frame) Exception) {
	*fm.defers = append(*fm.defers, f)
}

func (fm *Frame) runDefers() Exception {
	var exc Exception
	defers := *fm.defers
	for i := len(defers) - 1; i >= 0; i-- {
		exc2 := defers[i](fm)
		// TODO: Combine exc and exc2 if both are not nil
		if exc2 != nil && exc == nil {
			exc = exc2
		}
	}
	return exc
}

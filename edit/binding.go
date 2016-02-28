package edit

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// Errors thrown to Evaler.
var (
	ErrTakeNoArg       = errors.New("editor builtins take no arguments")
	ErrEditorInactive  = errors.New("editor inactive")
	ErrKeyMustBeString = errors.New("key must be string")
)

// BindingTable adapts a binding table to eval.IndexSetter, so that it can be
// manipulated in elvish script.
type BindingTable struct {
	inner map[Key]BoundFunc
}

func (BindingTable) Kind() string {
	return "map"
}

func (bt BindingTable) Repr(indent int) string {
	var builder eval.MapReprBuilder
	builder.Indent = indent
	for k, v := range bt.inner {
		builder.WritePair(parse.Quote(k.String()), v.Repr(eval.IncIndent(indent, 1)))
	}
	return builder.String()
}

func (bt BindingTable) IndexOne(idx eval.Value) eval.Value {
	key := keyIndex(idx)
	switch f := bt.inner[key].(type) {
	case Builtin:
		return eval.String(f.name)
	case CallerBoundFunc:
		return f.Caller
	}
	throw(errors.New("bug"))
	panic("unreachable")
}

func (bt BindingTable) IndexSet(idx, v eval.Value) {
	key := keyIndex(idx)

	var f BoundFunc
	switch v := v.(type) {
	case eval.String:
		builtin, ok := builtinMap[string(v)]
		if !ok {
			throw(fmt.Errorf("no builtin named %s", v.Repr(eval.NoPretty)))
		}
		f = builtin
	case eval.CallerValue:
		f = CallerBoundFunc{v}
	default:
		throw(fmt.Errorf("bad function type %s", v.Kind()))
	}

	bt.inner[key] = f
}

func keyIndex(idx eval.Value) Key {
	skey, ok := idx.(eval.String)
	if !ok {
		throw(ErrKeyMustBeString)
	}
	key, err := parseKey(string(skey))
	if err != nil {
		throw(err)
	}
	return key
}

// BuiltinCallerValue adapts a Builtin to satisfy eval.Value and eval.Caller,
// so that it can be called from elvish script.
type BuiltinCallerValue struct {
	b  Builtin
	ed *Editor
}

func (*BuiltinCallerValue) Kind() string {
	return "fn"
}

func (eb *BuiltinCallerValue) Repr(int) string {
	return "<editor builtin " + eb.b.name + ">"
}

func (eb *BuiltinCallerValue) Call(ec *eval.EvalCtx, args []eval.Value) {
	if len(args) > 0 {
		throw(ErrTakeNoArg)
	}
	if !eb.ed.active {
		throw(ErrEditorInactive)
	}
	eb.b.impl(eb.ed)
}

// BoundFunc is a function bound to a key. It is either a Builtin or an
// EvalCaller.
type BoundFunc interface {
	eval.Reprer
	Call(ed *Editor)
}

func (b Builtin) Repr(int) string {
	return b.name
}

func (b Builtin) Call(ed *Editor) {
	b.impl(ed)
}

// CallerBoundFunc adapts eval.Caller to BoundFunc, so that functions in elvish
// script can be bound to keys.
type CallerBoundFunc struct {
	Caller eval.CallerValue
}

func (c CallerBoundFunc) Repr(indent int) string {
	return c.Caller.Repr(indent)
}

func (c CallerBoundFunc) Call(ed *Editor) {
	rout, chanOut, ports, err := makePorts()
	if err != nil {
		return
	}

	// Goroutines to collect output.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		rd := bufio.NewReader(rout)
		for {
			line, err := rd.ReadString('\n')
			if err != nil {
				break
			}
			// XXX notify is not concurrency-safe.
			ed.notify("[bound fn bytes] %s", line[:len(line)-1])
		}
		rout.Close()
		wg.Done()
	}()
	go func() {
		for v := range chanOut {
			ed.notify("[bound fn value] %s", v.Repr(eval.NoPretty))
		}
		wg.Done()
	}()

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor]", "", ports)
	ex := ec.PCall(c.Caller, []eval.Value{})
	if ex != nil {
		ed.notify("function error: %s", ex.Error())
	}

	eval.ClosePorts(ports)
	wg.Wait()
	ed.refresh(true, true)
}

// makePorts connects stdin to /dev/null and a closed channel, identifies
// stdout and stderr and connects them to a pipe and channel. It returns the
// other end of stdout and the resulting []*eval.Port. The caller is
// responsible for closing the returned file and calling eval.ClosePorts on the
// ports.
func makePorts() (*os.File, chan eval.Value, []*eval.Port, error) {
	in, err := makeClosedStdin()
	if err != nil {
		return nil, nil, nil, err
	}

	// Output
	rout, out, err := os.Pipe()
	if err != nil {
		Logger.Println(err)
		return nil, nil, nil, err
	}
	chanOut := make(chan eval.Value)

	return rout, chanOut, []*eval.Port{
		in,
		{File: out, CloseFile: true, Chan: chanOut, CloseChan: true},
		{File: out, Chan: chanOut},
	}, nil
}

func makeClosedStdin() (*eval.Port, error) {
	// Input
	devnull, err := os.Open("/dev/null")
	if err != nil {
		Logger.Println(err)
		return nil, err
	}
	in := make(chan eval.Value)
	close(in)
	return &eval.Port{File: devnull, CloseFile: true, Chan: in}, nil
}

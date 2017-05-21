package edit

import (
	"bufio"
	"errors"
	"os"
	"sync"

	"github.com/elves/elvish/eval"
)

// This file implements types and functions for interactions with the
// Elvishscript runtime.

var (
	errNotNav         = errors.New("not in navigation mode")
	errMustBeString   = errors.New("must be string")
	errEditorInactive = errors.New("editor inactive")
)

// BuiltinFn records an editor builtin.
type BuiltinFn struct {
	name string
	impl func(ed *Editor)
}

var _ eval.CallableValue = &BuiltinFn{}

func (*BuiltinFn) Kind() string {
	return "fn"
}

func (bf *BuiltinFn) Repr(int) string {
	return "$" + bf.name
}

func (bf *BuiltinFn) Call(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	eval.TakeNoOpt(opts)
	eval.TakeNoArg(args)
	ed, ok := ec.Editor.(*Editor)
	if !ok || !ed.active {
		throw(errEditorInactive)
	}
	bf.impl(ed)
}

// installModules installs le: and le:* modules.
func installModules(modules map[string]eval.Namespace, ed *Editor) {
	// Construct the le: module.
	ns := eval.Namespace{}
	// Populate builtins.
	for _, b := range builtinMaps[""] {
		ns[eval.FnPrefix+b.name] = eval.NewPtrVariable(b)
	}

	// Populate binding tables in the variable $binding.
	// TODO Make binding specific to the Editor.
	binding := &eval.Struct{
		[]string{"insert", "command", "completion", "navigation", "history"},
		[]eval.Variable{
			eval.NewRoVariable(BindingTable{keyBindings[modeInsert]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeCommand]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeCompletion]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeNavigation]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeHistory]}),
		},
	}
	ns["binding"] = eval.NewRoVariable(binding)

	ns["completer"] = argCompleter
	ns[eval.FnPrefix+"complete-getopt"] = eval.NewRoVariable(
		&eval.BuiltinFn{"le:&complete-getopt", complGetopt})
	for _, bac := range argCompletersData {
		ns[eval.FnPrefix+bac.name] = eval.NewRoVariable(bac)
	}

	ns["prompt"] = ed.prompt
	ns["rprompt"] = ed.rprompt
	ns["rprompt-persistent"] = ed.rpromptPersistent
	ns["history"] = eval.NewRoVariable(History{&ed.historyMutex, ed.store})

	ns["current-command"] = eval.MakeVariableFromCallback(
		func(v eval.Value) {
			if !ed.active {
				throw(errEditorInactive)
			}
			if s, ok := v.(eval.String); ok {
				ed.line = string(s)
				ed.dot = len(ed.line)
			} else {
				throw(errMustBeString)
			}
		},
		func() eval.Value { return eval.String(ed.line) },
	)
	ns["selected-file"] = eval.MakeRoVariableFromCallback(
		func() eval.Value {
			if !ed.active {
				throw(errEditorInactive)
			}
			if ed.mode.Mode() != modeNavigation {
				throw(errNotNav)
			}
			return eval.String(ed.navigation.current.selectedName())
		},
	)

	ns["abbr"] = eval.NewRoVariable(eval.MapStringString(ed.abbreviations))

	ns["loc-pinned"] = ed.locationPinned
	ns["loc-hidden"] = ed.locationHidden

	ns["before-readline"] = ed.beforeReadLine
	ns["after-readline"] = ed.afterReadLine

	ns[eval.FnPrefix+"styled"] = eval.NewRoVariable(&eval.BuiltinFn{"le:&styled", styledBuiltin})

	modules["le"] = ns

	// Install other modules.
	for module, builtins := range builtinMaps {
		if module != "" {
			modules["le:"+module] = makeNamespaceFromBuiltins(builtins)
		}
	}
}

// CallFn calls an Fn, displaying its outputs and possible errors as editor
// notifications. It is the preferred way to call a Fn while the editor is
// active.
func (ed *Editor) CallFn(fn eval.CallableValue, args ...eval.Value) {
	if b, ok := fn.(*BuiltinFn); ok {
		// Builtin function: quick path.
		b.impl(ed)
		return
	}

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
			ed.Notify("[bytes output] %s", line[:len(line)-1])
		}
		rout.Close()
		wg.Done()
	}()
	go func() {
		for v := range chanOut {
			ed.Notify("[value output] %s", v.Repr(eval.NoPretty))
		}
		wg.Done()
	}()

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor]", "", ports)
	ex := ec.PCall(fn, args, eval.NoOpts)
	if ex != nil {
		ed.Notify("function error: %s", ex.Error())
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
	// Output
	rout, out, err := os.Pipe()
	if err != nil {
		Logger.Println(err)
		return nil, nil, nil, err
	}
	chanOut := make(chan eval.Value)

	return rout, chanOut, []*eval.Port{
		eval.DevNullClosedChan,
		{File: out, CloseFile: true, Chan: chanOut, CloseChan: true},
		{File: out, Chan: chanOut},
	}, nil
}

// callPrompt calls a Fn, assuming that it is a prompt. It calls the Fn with no
// arguments and closed input, and converts its outputs to styled objects.
func callPrompt(ed *Editor, fn eval.Callable) []*styled {
	ports := []*eval.Port{eval.DevNullClosedChan, &eval.Port{File: os.Stdout}, &eval.Port{File: os.Stderr}}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor prompt]", "", ports)
	values, err := ec.PCaptureOutput(fn, nil, eval.NoOpts)
	if err != nil {
		ed.Notify("prompt function error: %v", err)
		return nil
	}

	var ss []*styled
	for _, v := range values {
		if s, ok := v.(*styled); ok {
			ss = append(ss, s)
		} else {
			ss = append(ss, &styled{eval.ToString(v), styles{}})
		}
	}
	return ss
}

// callArgCompleter calls a Fn, assuming that it is an arg completer. It calls
// the Fn with specified arguments and closed input, and converts its output to
// candidate objects.
func callArgCompleter(fn eval.CallableValue, ev *eval.Evaler, words []string) ([]*candidate, error) {
	// Quick path for builtin arg completers.
	if builtin, ok := fn.(*builtinArgCompleter); ok {
		return builtin.impl(words, ev)
	}

	ports := []*eval.Port{eval.DevNullClosedChan, &eval.Port{File: os.Stdout}, &eval.Port{File: os.Stderr}}

	args := make([]eval.Value, len(words))
	for i, word := range words {
		args[i] = eval.String(word)
	}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ev, "[editor completer]", "", ports)
	values, err := ec.PCaptureOutput(fn, args, eval.NoOpts)
	if err != nil {
		return nil, errors.New("completer error: " + err.Error())
	}

	cands := make([]*candidate, len(values))
	for i, v := range values {
		switch v := v.(type) {
		case eval.String:
			cands[i] = newPlainCandidate(string(v))
		case *candidate:
			cands[i] = v
		default:
			return nil, errors.New("completer must output string or candidate")
		}
	}
	return cands, nil
}

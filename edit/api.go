package edit

import (
	"bufio"
	"errors"
	"os"
	"sync"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// This file implements types and functions for interactions with the
// Elvishscript runtime.

var (
	errNotNav         = errors.New("not in navigation mode")
	errMustBeString   = errors.New("must be string")
	errEditorInvalid  = errors.New("internal error: editor not set up")
	errEditorInactive = errors.New("editor inactive")
)

// BuiltinFn records an editor builtin.
type BuiltinFn struct {
	name string
	impl func(ed *Editor)
}

var _ eval.CallableValue = &BuiltinFn{}

// Kind returns "fn".
func (*BuiltinFn) Kind() string {
	return "fn"
}

// Eq compares based on identity.
func (bf *BuiltinFn) Eq(a interface{}) bool {
	return bf == a
}

// Repr returns the representation of a builtin function as a variable name.
func (bf *BuiltinFn) Repr(int) string {
	return "$" + bf.name
}

// Call calls a builtin function.
func (bf *BuiltinFn) Call(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	eval.TakeNoOpt(opts)
	eval.TakeNoArg(args)
	ed, ok := ec.Editor.(*Editor)
	if !ok {
		throw(errEditorInvalid)
	}
	if !ed.active {
		throw(errEditorInactive)
	}
	bf.impl(ed)
}

// installModules installs edit: and edit:* modules.
func installModules(modules map[string]eval.Namespace, ed *Editor) {
	// Construct the edit: module, starting with builtins.
	ns := makeNamespaceFromBuiltins(builtinMaps[""])

	// TODO(xiaq): Everything here should be registered to some registry instead
	// of centralized here.

	// Editor configurations.
	for name, variable := range ed.variables {
		ns[name] = variable
	}

	// Internal states.
	ns["history"] = eval.NewRoVariable(History{&ed.historyMutex, ed.daemon})
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
			nav, ok := ed.mode.(*navigation)
			if !ok {
				throw(errNotNav)
			}
			return eval.String(nav.current.selectedName())
		},
	)

	// Completers.
	for _, bac := range argCompletersData {
		ns[eval.FnPrefix+bac.name] = eval.NewRoVariable(bac)
	}

	// Matchers.
	eval.AddBuiltinFns(ns, matchers...)

	// Functions.
	eval.AddBuiltinFns(ns,
		&eval.BuiltinFn{"edit:command-history", CommandHistory},
		&eval.BuiltinFn{"edit:complete-getopt", complGetopt},
		&eval.BuiltinFn{"edit:complex-candidate", outputComplexCandidate},
		&eval.BuiltinFn{"edit:insert-at-dot", InsertAtDot},
		&eval.BuiltinFn{"edit:replace-input", ReplaceInput},
		&eval.BuiltinFn{"edit:styled", styled},
		&eval.BuiltinFn{"edit:wordify", Wordify},
		&eval.BuiltinFn{"edit:-dump-buf", _dumpBuf},
		&eval.BuiltinFn{"edit:-narrow-read", NarrowRead},
	)

	modules["edit"] = ns
	// Install other modules.
	for module, builtins := range builtinMaps {
		if module != "" {
			modules["edit:"+module] = makeNamespaceFromBuiltins(builtins)
		}
	}

	// Add $edit:xxx:binding variables.
	// TODO Make binding specific to the Editor.
	for _, mode := range []string{
		modeInsert, modeCommand, modeCompletion, modeNavigation, modeHistory,
		modeHistoryListing, modeLocation, modeLastCmd, modeListing, modeNarrow} {

		if modules["edit:"+mode] == nil {
			modules["edit:"+mode] = make(eval.Namespace)
		}
		modules["edit:"+mode]["binding"] =
			eval.NewRoVariable(BindingTable{keyBindings[mode]})
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
		logger.Println(err)
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
func callPrompt(ed *Editor, fn eval.Callable) []*ui.Styled {
	ports := []*eval.Port{eval.DevNullClosedChan, {File: os.Stdout}, {File: os.Stderr}}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor prompt]", "", ports)
	values, err := ec.PCaptureOutput(fn, nil, eval.NoOpts)
	if err != nil {
		ed.Notify("prompt function error: %v", err)
		return nil
	}

	var ss []*ui.Styled
	for _, v := range values {
		if s, ok := v.(*ui.Styled); ok {
			ss = append(ss, s)
		} else {
			ss = append(ss, &ui.Styled{eval.ToString(v), ui.Styles{}})
		}
	}
	return ss
}

// callArgCompleter calls a Fn, assuming that it is an arg completer. It calls
// the Fn with specified arguments and closed input, and converts its output to
// candidate objects.
func callArgCompleter(fn eval.CallableValue,
	ev *eval.Evaler, words []string) ([]rawCandidate, error) {

	// Quick path for builtin arg completers.
	if builtin, ok := fn.(*builtinArgCompleter); ok {
		return builtin.impl(words, ev)
	}

	ports := []*eval.Port{
		eval.DevNullClosedChan, {File: os.Stdout}, {File: os.Stderr}}

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

	cands := make([]rawCandidate, len(values))
	for i, v := range values {
		switch v := v.(type) {
		case rawCandidate:
			cands[i] = v
		case eval.String:
			cands[i] = plainCandidate(v)
		default:
			return nil, errors.New("completer must output string or candidate")
		}
	}
	return cands, nil
}

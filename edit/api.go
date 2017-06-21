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

// installModules installs le: and le:* modules.
func installModules(modules map[string]eval.Namespace, ed *Editor) {
	// Construct the le: module, starting with builtins.
	ns := makeNamespaceFromBuiltins(builtinMaps[""])

	// Populate binding tables in the variable $binding.
	// TODO Make binding specific to the Editor.
	binding := &eval.Struct{
		[]string{"insert", "command", "completion", "navigation", "history", "histlist"},
		[]eval.Variable{
			eval.NewRoVariable(BindingTable{keyBindings[modeInsert]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeCommand]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeCompletion]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeNavigation]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeHistory]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeHistoryListing]}),
		},
	}
	ns["binding"] = eval.NewRoVariable(binding)

	ns[eval.FnPrefix+"complete-getopt"] = eval.NewRoVariable(
		&eval.BuiltinFn{"le:&complete-getopt", complGetopt})
	ns[eval.FnPrefix+"complex-candidate"] = eval.NewRoVariable(
		&eval.BuiltinFn{"le:&complex-candidate", outputComplexCandidate})
	for _, bac := range argCompletersData {
		ns[eval.FnPrefix+bac.name] = eval.NewRoVariable(bac)
	}

	// Pour variables into the le: namespace.
	for name, variable := range ed.variables {
		ns[name] = variable
	}

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
			if ed.mode.Mode() != modeNavigation {
				throw(errNotNav)
			}
			return eval.String(ed.navigation.current.selectedName())
		},
	)

	ns[eval.FnPrefix+"styled"] = eval.NewRoVariable(&eval.BuiltinFn{"le:&styled", styled})

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

// outputComplexCandidate composes a complexCandidate from its args.
func outputComplexCandidate(ec *eval.EvalCtx, a []eval.Value, o map[string]eval.Value) {
	var style string

	c := &complexCandidate{}

	eval.ScanArgs(a, &c.stem)
	eval.ScanOpts(o,
		eval.Opt{"code-suffix", &c.codeSuffix, eval.String("")},
		eval.Opt{"display-suffix", &c.displaySuffix, eval.String("")},
		eval.Opt{"style", &style, eval.String("")},
	)
	if style != "" {
		c.style = ui.StylesFromString(style)
	}

	out := ec.OutputChan()
	out <- c
}

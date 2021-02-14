// Package edit implements the line editor for Elvish.
//
// The line editor is based on the cli package, which implements a general,
// Elvish-agnostic line editor, and multiple "addon" packages. This package
// glues them together and provides Elvish bindings for them.
package edit

import (
	"fmt"
	"sync"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/store"
)

// Editor is the interactive line editor for Elvish.
type Editor struct {
	app cli.App
	ns  *eval.Ns

	excMutex sync.RWMutex
	excList  vals.List
}

// An interface that wraps notifyf and notifyError. It is only implemented by
// the *Editor type; functions may take a notifier instead of *Editor argument
// to make it clear that they do not depend on other parts of *Editor.
type notifier interface {
	notifyf(format string, args ...interface{})
	notifyError(ctx string, e error)
}

// NewEditor creates a new editor. The TTY is used for input and output. The
// Evaler is used for syntax highlighting, completion, and calling callbacks.
// The Store is used for saving and retrieving command and directory history.
func NewEditor(tty cli.TTY, ev *eval.Evaler, st store.Store) *Editor {
	// Declare the Editor with a nil App first; some initialization functions
	// require a notifier as an argument, but does not use it immediately.
	ed := &Editor{excList: vals.EmptyList}
	nb := eval.NsBuilder{}
	appSpec := cli.AppSpec{TTY: tty}

	hs, err := newHistStore(st)
	if err != nil {
		// TODO(xiaq): Report the error.
	}

	initHighlighter(&appSpec, ev)
	initMaxHeight(&appSpec, nb)
	initReadlineHooks(&appSpec, ev, nb)
	initAddCmdFilters(&appSpec, ev, nb, hs)
	initGlobalBindings(&appSpec, ed, ev, nb)
	initInsertAPI(&appSpec, ed, ev, nb)
	initPrompts(&appSpec, ed, ev, nb)
	ed.app = cli.NewApp(appSpec)

	initExceptionsAPI(ed, nb)
	initVarsAPI(ed, nb)
	initCommandAPI(ed, ev, nb)
	initListings(ed, ev, st, hs, nb)
	initNavigation(ed, ev, nb)
	initCompletion(ed, ev, nb)
	initHistWalk(ed, ev, hs, nb)
	initInstant(ed, ev, nb)
	initMinibuf(ed, ev, nb)

	initBufferBuiltins(ed.app, nb)
	initTTYBuiltins(ed.app, tty, nb)
	initMiscBuiltins(ed.app, nb)
	initStateAPI(ed.app, nb)
	initStoreAPI(ed.app, nb, hs)

	ed.ns = nb.Ns()
	evalDefaultBinding(ev, ed.ns)

	return ed
}

//elvdoc:var exceptions
//
// A list of exceptions thrown from callbacks such as prompts. Useful for
// examining tracebacks and other metadata.

func initExceptionsAPI(ed *Editor, nb eval.NsBuilder) {
	nb.Add("exceptions", vars.FromPtrWithMutex(&ed.excList, &ed.excMutex))
}

func evalDefaultBinding(ev *eval.Evaler, ns *eval.Ns) {
	src := parse.Source{Name: "[default bindings]", Code: defaultBindingsElv}
	err := ev.Eval(src, eval.EvalCfg{Global: ns})
	if err != nil {
		panic(err)
	}
}

// ReadCode reads input from the user.
func (ed *Editor) ReadCode() (string, error) {
	return ed.app.ReadCode()
}

// Ns returns a namespace for manipulating the editor from Elvish code.
//
// See https://elv.sh/ref/edit.html for the Elvish API.
func (ed *Editor) Ns() *eval.Ns {
	return ed.ns
}

func (ed *Editor) notifyf(format string, args ...interface{}) {
	ed.app.Notify(fmt.Sprintf(format, args...))
}

func (ed *Editor) notifyError(ctx string, e error) {
	if exc, ok := e.(eval.Exception); ok {
		ed.excMutex.Lock()
		defer ed.excMutex.Unlock()
		ed.excList = ed.excList.Cons(exc)
		ed.notifyf("[%v error] %v\n"+
			`see stack trace with "show $edit:exceptions[%d]"`,
			ctx, e, ed.excList.Len()-1)
	} else {
		ed.notifyf("[%v error] %v", ctx, e)
	}
}

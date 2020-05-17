// Package edit implements the line editor for Elvish.
//
// The line editor is based on the cli package, which implements a general,
// Elvish-agnostic line editor, and multiple "addon" packages. This package
// glues them together and provides Elvish bindings for them.
package edit

import (
	"fmt"
	"os"
	"sync"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/store"
)

// Editor is the interface line editor for Elvish.
type Editor struct {
	app cli.App
	ns  eval.Ns

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

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(tty cli.TTY, ev *eval.Evaler, st store.Store) *Editor {
	// Declare the Editor with a nil App first; some initialization functions
	// require a notifier as an argument, but does not use it immediately.
	ed := &Editor{ns: eval.Ns{}, excList: vals.EmptyList}
	appSpec := cli.AppSpec{TTY: tty}

	fuser, err := histutil.NewFuser(st)
	if err != nil {
		// TODO(xiaq): Report the error.
	}

	initHighlighter(&appSpec, ev)
	initMaxHeight(&appSpec, ed.ns)
	initReadlineHooks(&appSpec, ev, ed.ns)
	if fuser != nil {
		initAddCmdFilters(&appSpec, ev, ed.ns, fuser)
	}
	initInsertAPI(&appSpec, ed, ev, ed.ns)
	initPrompts(&appSpec, ed, ev, ed.ns)
	ed.app = cli.NewApp(appSpec)

	initExceptionsAPI(ed)
	initCommandAPI(ed, ev)
	initListings(ed, ev, st, fuser)
	initNavigation(ed, ev)
	initCompletion(ed, ev)
	initHistWalk(ed, ev, fuser)
	initInstant(ed, ev)
	initMinibuf(ed, ev)

	initBufferBuiltins(ed.app, ed.ns)
	initTTYBuiltins(ed.app, tty, ed.ns)
	initMiscBuiltins(ed.app, ed.ns)
	initStateAPI(ed.app, ed.ns)
	initStoreAPI(ed.app, ed.ns, fuser)
	evalDefaultBinding(ev, ed.ns)

	return ed
}

//elvdoc:var exceptions
//
// A list of exceptions thrown from callbacks such as prompts. Useful for
// examining tracebacks and other metadata.

func initExceptionsAPI(ed *Editor) {
	ed.ns.Add("exceptions", vars.FromPtrWithMutex(&ed.excList, &ed.excMutex))
}

func evalDefaultBinding(ev *eval.Evaler, ns eval.Ns) {
	// TODO(xiaq): The evaler API should accodomate the use case of evaluating a
	// piece of code in an alternative global namespace.

	src := parse.Source{Name: "[default bindings]", Code: defaultBindingsElv}
	tree, err := parse.Parse(src)
	if err != nil {
		panic(err)
	}
	op, err := ev.CompileWithGlobal(tree, ns, nil)
	if err != nil {
		panic(err)
	}
	// TODO(xiaq): Use stdPorts when it is possible to do so.
	fm := eval.NewTopFrame(ev, src, []*eval.Port{
		{File: os.Stdin}, {File: os.Stdout}, {File: os.Stderr},
	})
	fm.SetLocal(ns)
	err = fm.Eval(op)
	if err != nil {
		panic(err)
	}
}

// ReadCode reads input from the user.
func (ed *Editor) ReadCode() (string, error) {
	return ed.app.ReadCode()
}

// Ns returns a namespace for manipulating the editor from Elvish code.
func (ed *Editor) Ns() eval.Ns {
	return ed.ns
}

func (ed *Editor) notify(s string) { ed.app.Notify(s) }

func (ed *Editor) notifyf(format string, args ...interface{}) {
	ed.app.Notify(fmt.Sprintf(format, args...))
}

func (ed *Editor) notifyError(ctx string, e error) {
	if exc, ok := e.(*eval.Exception); ok {
		ed.excMutex.Lock()
		defer ed.excMutex.Unlock()
		ed.excList = ed.excList.Cons(exc)
		ed.notifyf("[%v error] %v\n"+
			`see stack trace with "use exc; exc:show $edit:exceptions[%d]"`,
			ctx, e, ed.excList.Len()-1)
	} else {
		ed.notifyf("[%v error] %v", ctx, e)
	}
}

package cliedit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cliedit/highlight"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store/storedefs"
)

// Editor is the interface line editor for Elvish.
//
// This currently implements the same interface as *Editor in the old edit
// package to ease transition.
//
// TODO: Rename ReadLine to ReadCode and remove Close.
type Editor struct {
	app *cli.App
	ns  eval.Ns
}

// Wraps the histutil.Fuser interface to implement histutil.Store. This is a
// bandaid as we cannot change the implementation of Fuser without breaking its
// other users. Eventually Fuser should implement Store directly.
type fuserWrapper struct {
	*histutil.Fuser
}

func (f fuserWrapper) AddCmd(cmd histutil.Entry) (int, error) {
	return f.Fuser.AddCmd(cmd.Text)
}

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(in, out *os.File, ev *eval.Evaler, st storedefs.Store) *Editor {
	ns := eval.NewNs()
	cfg := &cli.AppConfig{}
	// TODO: Remove the forward declaration. Currently this is needed by
	// makePrompt only.
	var app *cli.App

	cfg.Highlighter = highlight.NewHighlighter(
		highlight.Dep{Check: makeCheck(ev), HasCommand: makeHasCommand(ev)})

	histFuser, err := histutil.NewFuser(st)
	if err == nil {
		_ = histFuser
		cfg.HistoryStore = fuserWrapper{histFuser}
	} else {
		fmt.Fprintln(out, "failed to initialize history facilities")
	}

	ns.Add("max-height", vars.FromPtr(&cfg.MaxHeight))

	// TODO: BindingMap should pass event context to event handlers
	ns.AddGoFns("<edit>", map[string]interface{}{
		"binding-map": makeBindingMap,
		"commit-code": cli.CommitCode,
		"commit-eof":  cli.CommitEOF,
		"reset-mode":  cli.ResetMode,
	}).AddGoFns("<edit>", bufferBuiltins)

	// Elvish hook APIs
	var beforeReadline func()
	ns["before-readline"], beforeReadline = initBeforeReadline(ev)
	var afterReadline func(string)
	ns["after-readline"], afterReadline = initAfterReadline(ev)
	cfg.BeforeReadline = []func(){beforeReadline}
	cfg.AfterReadline = []func(string){afterReadline}

	// Prompts
	cfg.Prompt = makePrompt(app, ev, ns, defaultPrompt, "prompt")
	cfg.RPrompt = makePrompt(app, ev, ns, defaultRPrompt, "rprompt")

	// Insert mode
	insertNs := initInsert(ev, &cfg.InsertModeConfig)
	ns.AddNs("insert", insertNs)

	// Listing modes.
	lsBinding, lsNs := initListing()
	ns.AddNs("listing", lsNs)

	lastcmdNs := initLastcmd(ev, lsBinding, &cfg.LastcmdModeConfig)
	ns.AddNs("lastcmd", lastcmdNs)

	histlistNs := initHistlist(ev, lsBinding, &cfg.HistlistModeConfig)
	ns.AddNs("histlist", histlistNs)

	// Evaluate default bindings.
	evalDefaultBinding(ev, ns)

	app = cli.NewAppFromFiles(cfg, in, out)
	return &Editor{app, ns}
}

func evalDefaultBinding(ev *eval.Evaler, ns eval.Ns) {
	// TODO(xiaq): The evaler API should accodomate the use case of evaluating a
	// piece of code in an alternative global namespace.

	n, err := parse.AsChunk("[default bindings]", defaultBindingsElv)
	if err != nil {
		panic(err)
	}
	src := eval.NewScriptSource(
		"[default bindings]", "[default bindings]", defaultBindingsElv)
	op, err := ev.CompileWithGlobal(n, src, ns)
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

// ReadLine reads input from the user.
func (ed *Editor) ReadLine() (string, error) {
	return ed.app.ReadCode()
}

// Ns returns a namespace for manipulating the editor from Elvish code.
func (ed *Editor) Ns() eval.Ns {
	return ed.ns
}

// Close is a no-op.
func (ed *Editor) Close() {}

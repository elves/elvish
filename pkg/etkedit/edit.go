package edit

import (
	_ "embed"
	"sync"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/ui"
)

type Editor struct {
	ev *eval.Evaler
	ns *eval.Ns

	store     storedefs.Store
	histStore *histStore

	mutex sync.RWMutex

	etkCtx *etk.Context

	// Global hooks
	afterCommand   vals.List
	beforeReadline vals.List
	afterReadline  vals.List
	// Global app config
	maxHeight int
	// Key bindings
	globalBinding     bindingsMap
	insertBinding     bindingsMap
	minibufBinding    bindingsMap
	instantBinding    bindingsMap
	commandBinding    bindingsMap
	completionBinding bindingsMap
	navigationBinding bindingsMap
	locationBinding   bindingsMap
	histlistBinding   bindingsMap
	historyBinding    bindingsMap
	lastcmdBinding    bindingsMap
	// Main TextArea: prompts
	prompt  promptCfg
	rprompt promptCfg
	// Main TextArea: abbreviations
	simpleAbbr    vals.Map
	commandAbbr   vals.Map
	smallWordAbbr vals.Map
}

var (
	// TODO: Use # for root
	defaultPromptFn = func() ui.Text { return ui.T(fsutil.Getwd() + "> ") }
	// TODO: Use real username and hostname
	defaultRPromptFn = func() ui.Text { return ui.T("user@host", ui.Inverse) }
)

//go:embed init.elv
var initElv string

func NewEditor(ev *eval.Evaler, st storedefs.Store) *Editor {
	hs, err := newHistStore(st)
	if err != nil {
		// TODO: Handle the error
		_ = err
	}

	ed := &Editor{
		ev: ev,

		store:     st,
		histStore: hs,

		afterCommand:   vals.EmptyList,
		beforeReadline: vals.EmptyList,
		afterReadline:  vals.EmptyList,

		maxHeight: 0,

		globalBinding:     emptyBindingsMap,
		insertBinding:     emptyBindingsMap,
		minibufBinding:    emptyBindingsMap,
		instantBinding:    emptyBindingsMap,
		commandBinding:    emptyBindingsMap,
		completionBinding: emptyBindingsMap,
		navigationBinding: emptyBindingsMap,
		locationBinding:   emptyBindingsMap,
		histlistBinding:   emptyBindingsMap,
		historyBinding:    emptyBindingsMap,
		lastcmdBinding:    emptyBindingsMap,

		prompt:  makeDefaultPromptCfg(func() ui.Text { return defaultPromptFn() }),
		rprompt: makeDefaultPromptCfg(func() ui.Text { return defaultRPromptFn() }),

		simpleAbbr:    vals.EmptyMap,
		commandAbbr:   vals.EmptyMap,
		smallWordAbbr: vals.EmptyMap,
	}

	// Build the namespace.
	ed.ns = eval.BuildNsNamed("edit").
		// Global functions.
		AddGoFns(map[string]any{
			// Global state
			"close-mode": wrapCtxFn(ed, popAddon),
			// Main TextArea: buffer state
			"insert-at-dot": wrapCtxFn1(ed, insertAtDot),
			"replace-input": wrapCtxFn1(ed, replaceInput),
			// Helpers that happen to be in the edit: namespace, but don't
			// actually depend on Editor
			"binding-table": makeBindingMap,
			"key":           toKey,
		}).
		AddGoFns(ed.codeBufferBuiltins()).
		// Global variables.
		AddVars(map[string]vars.Var{
			// Keep this in the same order as fields in Editor.
			// Global hooks
			"after-command":   makeEditVar(ed, &ed.afterCommand),
			"before-readline": makeEditVar(ed, &ed.beforeReadline),
			"after-readline":  makeEditVar(ed, &ed.afterReadline),
			// Global app config
			"max-height": makeEditVar(ed, &ed.maxHeight),
			// Key bindings
			"global-binding": makeEditVar(ed, &ed.globalBinding),
			// Main TextArea: prompts
			"prompt":                    makeEditVar(ed, &ed.prompt.Fn),
			"-prompt-eagerness":         makeEditVar(ed, &ed.prompt.Eagerness),
			"prompt-stale-threshold":    makeEditVar(ed, &ed.prompt.StaleThreshold),
			"prompt-stale-transformer":  makeEditVar(ed, &ed.prompt.StaleTransformer),
			"rprompt":                   makeEditVar(ed, &ed.rprompt.Fn),
			"-rprompt-eagerness":        makeEditVar(ed, &ed.rprompt.Eagerness),
			"rprompt-stale-threshold":   makeEditVar(ed, &ed.rprompt.StaleThreshold),
			"rprompt-stale-transformer": makeEditVar(ed, &ed.rprompt.StaleTransformer),
			// Main TextArea: abbreviations
			"abbr":            makeEditVar(ed, &ed.simpleAbbr),
			"command-abbr":    makeEditVar(ed, &ed.commandAbbr),
			"small-word-abbr": makeEditVar(ed, &ed.smallWordAbbr),

			// Main TextArea: buffer state
			"-dot": bufferFieldVar[int]{
				ed,
				func(buf comps.TextBuffer) int { return buf.Dot },
				func(buf comps.TextBuffer, dot int) comps.TextBuffer {
					return comps.TextBuffer{Content: buf.Content, Dot: dot}
				},
			},
			"current-command": bufferFieldVar[string]{
				ed,
				func(buf comps.TextBuffer) string { return buf.Content },
				func(_ comps.TextBuffer, content string) comps.TextBuffer {
					return comps.TextBuffer{Content: content, Dot: len(content)}
				},
			},
			// Ordinary variable to be used in init.elv
			"command-duration": vars.FromInit(0.0),
		}).
		// Per-addon API.
		AddNs("insert", eval.BuildNsNamed("edit:insert").
			AddVar("binding", makeEditVar(ed, &ed.insertBinding))).
		AddNs("minibuf", eval.BuildNsNamed("edit:minibuf").
			AddGoFn("start", wrapCtxFn(ed, startMinibuf)).
			AddVar("binding", makeEditVar(ed, &ed.minibufBinding))).
		AddNs("instant", eval.BuildNsNamed("edit:instant").
			AddGoFn("start", wrapCtxFn(ed, startInstant)).
			AddVar("binding", makeEditVar(ed, &ed.instantBinding))).
		AddNs("command", eval.BuildNsNamed("edit:command").
			AddGoFn("start", wrapCtxFn(ed, startCommand)).
			AddVar("binding", makeEditVar(ed, &ed.commandBinding))).
		AddNs("completion", eval.BuildNsNamed("edit:completion").
			AddGoFns(map[string]any{
				"start":      wrapCtxFnEd(ed, startCompletion),
				"up":         wrapListingSelect(ed, listingUp),
				"down":       wrapListingSelect(ed, listingDown),
				"up-cycle":   wrapListingSelect(ed, listingUpCycle),
				"down-cycle": wrapListingSelect(ed, listingDownCycle),
				"left":       wrapListingSelect(ed, listingLeft),
				"right":      wrapListingSelect(ed, listingRight),
			}).
			AddVar("binding", makeEditVar(ed, &ed.completionBinding))).
		AddNs("navigation", eval.BuildNsNamed("edit:navigation").
			AddGoFn("start", wrapCtxFnEd(ed, startNavigation)).
			AddVar("binding", makeEditVar(ed, &ed.navigationBinding))).
		AddNs("location", eval.BuildNsNamed("edit:location").
			AddGoFn("start", wrapCtxFnEd(ed, startLocation)).
			AddVar("binding", makeEditVar(ed, &ed.locationBinding))).
		AddNs("histlist", eval.BuildNsNamed("edit:histlist").
			AddGoFn("start", wrapCtxFnEd(ed, startHistlist)).
			AddVar("binding", makeEditVar(ed, &ed.histlistBinding))).
		AddNs("history", eval.BuildNsNamed("edit:history").
			AddGoFn("start", wrapCtxFnEd(ed, startHistwalk)).
			AddVar("binding", makeEditVar(ed, &ed.historyBinding))).
		AddNs("lastcmd", eval.BuildNsNamed("edit:lastcmd").
			AddGoFn("start", wrapCtxFnEd(ed, startLastcmd)).
			AddVar("binding", makeEditVar(ed, &ed.lastcmdBinding))).
		Ns()

	// Run the init script.
	src := parse.Source{Name: "[init.elv]", Code: initElv}
	err = ev.Eval(src, eval.EvalCfg{Global: ed.ns})
	if err != nil {
		panic(err)
	}

	return ed
}

func (ed *Editor) Ns() *eval.Ns {
	return ed.ns
}

func (ed *Editor) Comp() etk.Comp {
	return etk.ModComp(
		app,
		// Note: we don't use etkBindingFromBindingMap here because this
		// also handles the showing of the "unbound" message.
		etk.InitState("binding", etkBindingFromBindingMap(ed, &ed.globalBinding)),
		etk.InitState("code/binding", func(c etk.Context, ev term.Event) etk.Reaction {
			switch ev {
			case term.K('D', ui.Ctrl):
				// TODO: Move this to binding map and use
				// etkBindingFromBindingMap
				return etk.FinishEOF
			default:
				handled := ed.callBinding(&ed.insertBinding, ev)
				if handled {
					return etk.Consumed
				}
			}
			return etk.Unused
		}),
		// TODO: Break up this into smaller hooks.
		etk.BeforeHook("various", func(c etk.Context) {
			ed.mutex.RLock()
			defer ed.mutex.RUnlock()

			bufferVar := bufferVar(c)
			callPrompt(c, "code/prompt", ed.prompt, bufferVar)
			callPrompt(c, "code/rprompt", ed.rprompt, bufferVar)
			// These live in the WithBefore rather than WithInit, because we
			// want mutations to the edit binding variables (made via keybinding
			// or minibuf) to be reflected in the same event loop.
			c.Set("code/abbr", convertAbbr(ed.simpleAbbr))
			c.Set("code/command-abbr", convertAbbr(ed.commandAbbr))
			c.Set("code/small-word-abbr", convertAbbr(ed.smallWordAbbr))
		}),
	)
}

func (ed *Editor) ReadCode(tty cli.TTY) (string, error) {
	ed.callHook("$edit:before-readline", &ed.beforeReadline)
	tty.ResetBuffer() // TODO: This was easy to miss
	m, err := etk.Run(ed.Comp(), etk.RunCfg{
		TTY: tty, Frame: ed.ev.CallFrame("edit"), MaxHeight: ed.maxHeight,
		ContextFn: func(c etk.Context) {
			setField(ed, &ed.etkCtx, &c)
		},
	})
	setField(ed, &ed.etkCtx, nil)
	if err != nil {
		return "", err
	}
	// TODO: Multi-level indexing should be easier
	codeArea, _ := m.Index("code")
	buf, _ := codeArea.(vals.Map).Index("buffer")
	code := buf.(comps.TextBuffer).Content
	ed.callHook("$edit:after-readline", &ed.afterReadline, code)
	return code, nil
}

func (ed *Editor) RunAfterCommandHooks(src parse.Source, duration float64, err error) {
	ed.callHook("$edit:after-command", &ed.afterCommand,
		vals.MakeMap("src", src, "duration", duration, "error", err))
}

func (ed *Editor) callHook(name string, hookPtr *vals.List, args ...any) {
	// TODO: Don't use eval.CallHook.
	eval.CallHook(ed.ev, nil, name, getField(ed, hookPtr), args...)
}

func getField[T any](ed *Editor, fieldPtr *T) T {
	ed.mutex.RLock()
	defer ed.mutex.RUnlock()
	return *fieldPtr
}

func setField[T any](ed *Editor, fieldPtr *T, value T) {
	ed.mutex.Lock()
	defer ed.mutex.Unlock()
	*fieldPtr = value
}

// Creates an editVar. This has to be a function because methods can't be
// polymorphic.
func makeEditVar[F any](ed *Editor, ptr *F) editVar[F] {
	return editVar[F]{ptr, ed.ev, &ed.mutex}
}

// Like [vars.PtrVar], but supports scanning Elvish functions as Go functions.
type editVar[F any] struct {
	ptr   *F
	ev    *eval.Evaler
	mutex *sync.RWMutex
}

func (v editVar[F]) Get() any {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return *v.ptr
}

func (v editVar[F]) Set(val any) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	scanned, err := etk.ScanToGo[F](val, v.ev.CallFrame("edit"))
	if err != nil {
		return err
	}
	*v.ptr = scanned
	return nil
}

func convertAbbr(m vals.Map) func(func(a, f string)) {
	return func(f func(a, b string)) {
		for it := m.Iterator(); it.HasElem(); it.Next() {
			k, v := it.Elem()
			ks, kok := k.(string)
			vs, vok := v.(string)
			if !kok || !vok {
				continue
			}
			f(ks, vs)
		}
	}
}

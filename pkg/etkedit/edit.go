package edit

import (
	"fmt"
	"sync"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/ui"
)

type Editor struct {
	ev *eval.Evaler

	mutex sync.RWMutex

	prompt, rprompt                        promptCfg
	simpleAbbr, commandAbbr, smallWordAbbr vals.Map
	insertBinding                          bindingsMap
}

var (
	// TODO: Use # for root
	defaultPromptFn = func() ui.Text { return ui.T(fsutil.Getwd() + "> ") }
	// TODO: Use real username and hostname
	defaultRPromptFn = func() ui.Text { return ui.T("user@host", ui.Inverse) }
)

func NewEditor(ev *eval.Evaler) *Editor {
	return &Editor{
		ev:         ev,
		prompt:     makeDefaultPromptCfg(func() ui.Text { return defaultPromptFn() }),
		rprompt:    makeDefaultPromptCfg(func() ui.Text { return defaultRPromptFn() }),
		simpleAbbr: vals.EmptyMap, commandAbbr: vals.EmptyMap, smallWordAbbr: vals.EmptyMap,
		insertBinding: emptyBindingsMap,
	}
}

func (ed *Editor) Ns() *eval.Ns {
	return eval.BuildNsNamed("edit").
		AddGoFns(map[string]any{
			"binding-table": makeBindingMap,
			"key":           toKey,
		}).
		AddVars(map[string]vars.Var{
			"prompt":                   makeEditVar(ed, &ed.prompt.Fn),
			"-prompt-eagerness":        makeEditVar(ed, &ed.prompt.Eagerness),
			"prompt-stale-threshold":   makeEditVar(ed, &ed.prompt.StaleThreshold),
			"prompt-stale-transformer": makeEditVar(ed, &ed.prompt.StaleTransformer),

			"rprompt":                   makeEditVar(ed, &ed.rprompt.Fn),
			"-rprompt-eagerness":        makeEditVar(ed, &ed.rprompt.Eagerness),
			"rprompt-stale-threshold":   makeEditVar(ed, &ed.rprompt.StaleThreshold),
			"rprompt-stale-transformer": makeEditVar(ed, &ed.rprompt.StaleTransformer),

			"abbr":            makeEditVar(ed, &ed.simpleAbbr),
			"command-abbr":    makeEditVar(ed, &ed.commandAbbr),
			"small-word-abbr": makeEditVar(ed, &ed.smallWordAbbr),
		}).
		AddNs("insert", eval.BuildNsNamed("edit:insert").
			AddVar("binding", makeEditVar(ed, &ed.insertBinding))).
		Ns()
}

func (ed *Editor) Comp() etk.Comp {
	return etk.WithBefore(
		etk.WithInit(
			App,
			"binding", func(ev term.Event, c etk.Context, r etk.React) etk.Reaction {
				reaction := r(ev)
				if reaction == etk.Unused {
					switch ev {
					case term.K('x', ui.Alt):
						PushAddon(c, etk.WithInit(etk.TextArea, "prompt", ui.T("minibuf> ")))
					default:
						if k, ok := ev.(term.KeyEvent); ok {
							c.AddMsg(ui.T(fmt.Sprintf("Unbound: %s", ui.Key(k))))
						}
					}
				}
				return reaction
			},
			"code/binding", func(ev term.Event, c etk.Context, r etk.React) etk.Reaction {
				reaction := r(ev)
				if reaction == etk.Unused {
					switch ev {
					case term.K('D', ui.Ctrl):
						return etk.FinishEOF
					}
				}
				return reaction
			},
		),
		func(c etk.Context) {
			ed.mutex.RLock()
			defer ed.mutex.RUnlock()

			bufferContent := etk.BindState(c, "code/buffer", etk.TextBuffer{}).Get().Content
			callPrompt(c, "code/prompt", ed.prompt, bufferContent)
			callPrompt(c, "code/rprompt", ed.rprompt, bufferContent)
			// These live in the WithBefore rather than WithInit, because we
			// want mutations to the edit binding variables (made via keybinding
			// or minibuf) to be reflected in the same event loop.
			c.Set("code/abbr", convertAbbr(ed.simpleAbbr))
			c.Set("code/command-abbr", convertAbbr(ed.commandAbbr))
			c.Set("code/small-word-abbr", convertAbbr(ed.smallWordAbbr))
		})
}

func (ed *Editor) ReadCode(tty cli.TTY) (string, error) {
	tty.ResetBuffer() // TODO: This was easy to miss
	m, err := etk.Run(tty, ed.ev.CallFrame("edit"), ed.Comp())
	if err != nil {
		return "", err
	}
	// TODO: Multi-level indexing should be easier
	codeArea, _ := m.Index("code")
	buf, _ := codeArea.(vals.Map).Index("buffer")
	return buf.(etk.TextBuffer).Content, nil
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

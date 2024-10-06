package edit_test

import (
	"embed"
	"testing"
	"time"

	"src.elv.sh/pkg/etk/etktest"
	edit "src.elv.sh/pkg/etkedit"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"edit-fixture", func(t *testing.T, ev *eval.Evaler) {
			// The default prompt and rprompt depends on environment factors
			// like the working directory. We don't need that for the majority
			// of tests.
			testutil.Set(t, edit.DefaultPromptFnPtr, constantly(ui.T("etkedit> ")))
			testutil.Set(t, edit.DefaultRPromptFnPtr, constantly(ui.T("")))

			// The real HasCommand implementation depends on the existence of
			// external commands, and may deliver results asynchronously depending
			// on how fast it is. For the majority of tests, we want deterministic
			// behavior instead.
			testutil.Set(t, edit.HasCommandPtr, hasBuiltinCmd)
			testutil.Set(t, edit.HasCommandMaxBlockPtr, func() time.Duration { return time.Hour })

			ed := edit.NewEditor(ev)
			ev.ExtendBuiltin(eval.BuildNs().AddNs("edit", ed))
			etktest.Setup(t, ev, ed.Comp())
		},
		"slow-blocked-has-command", func(t *testing.T, ev *eval.Evaler) {
			ch := make(chan struct{})
			ev.ExtendBuiltin(eval.BuildNs().AddGoFn("unblock-has-command", func() {
				close(ch)
			}))
			testutil.Set(t, edit.HasCommandPtr, func(ev *eval.Evaler, cmd string) bool {
				b := edit.HasCommandImpl(ev, cmd)
				<-ch
				return b
			})
			testutil.Set(t, edit.HasCommandMaxBlockPtr, func() time.Duration { return 0 })
		},
		"real-fast-has-command", func(t *testing.T) {
			testutil.Set(t, edit.HasCommandPtr, edit.HasCommandImpl)
			// edit-fixture has already set HasCommandMaxBlockPtr to an hour, so
			// the real HasCommand is fast relatively to that.
		},
	)
	/*
		"abbr-table-in-global", evaltest.GoFnInGlobal("abbr-table",
			func(m vals.Map) func(f func(a, f string)) {
				return func(f func(a, f string)) {
					for it := m.Iterator(); it.HasElem(); it.Next() {
						k, v := it.Elem()
						f(vals.ToString(k), vals.ToString(v))
					}
				}
			}),
	*/
}

func hasBuiltinCmd(ev *eval.Evaler, cmd string) bool {
	return eval.IsBuiltinSpecial[cmd] || ev.Builtin().HasKeyString(cmd+eval.FnSuffix)
}

func constantly[T any](v T) func() T { return func() T { return v } }

package newedit

import (
	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/clitypes"
	"github.com/elves/elvish/newedit/cliutil"
	"github.com/elves/elvish/newedit/insert"
	"github.com/xiaq/persistent/hashmap"
)

// Initializes states for the insert mode and its API.
func initInsert(ed editor, ev *eval.Evaler) (*insert.Mode, eval.Ns) {
	// Underlying abbreviation map and binding map.
	abbr := vals.EmptyMap
	binding := emptyBindingMap

	m := &insert.Mode{
		KeyHandler:  keyHandlerFromBindings(ed, ev, &binding),
		AbbrIterate: func(cb func(a, f string)) { abbrIterate(abbr, cb) },
	}

	st := ed.State()

	ns := eval.Ns{
		"binding": vars.FromPtr(&binding),
		"abbr":    vars.FromPtr(&abbr),
		"quote-paste": vars.FromPtrWithMutex(
			&m.Config.Raw.QuotePaste, &m.Config.Mutex),
	}.AddGoFns("<edit:insert>:", map[string]interface{}{
		"start": func() { st.SetMode(m) },
		"default-handler": func() error {
			action := cliutil.BasicHandler(tty.KeyEvent(st.BindingKey()), st)
			if action != clitypes.NoAction {
				return cliutil.ActionError(action)
			}
			return nil
		},
	})

	return m, ns
}

// Iterates through each pair in the hashmap and calls the callback for each
// pair, ignoring pairs that are mistyped.
func abbrIterate(abbr hashmap.Map, cb func(a, f string)) {
	for it := abbr.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		abbr, kok := k.(string)
		full, vok := v.(string)
		if !kok || !vok {
			continue
		}
		cb(abbr, full)
	}
}

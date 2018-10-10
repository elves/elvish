package newedit

import (
	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/insert"
	"github.com/elves/elvish/newedit/utils"
	"github.com/xiaq/persistent/hashmap"
)

// Initializes states for the insert mode and its API.
func initInsert(ed editor, ev *eval.Evaler) (*insert.Mode, eval.Ns) {
	abbr := vals.EmptyMap
	binding := EmptyBindingMap

	m := &insert.Mode{
		KeyHandler:  keyHandlerFromBinding(ed, ev, &binding),
		AbbrIterate: func(cb func(a, f string)) { abbrIterate(abbr, cb) },
	}

	ns := eval.Ns{
		"binding": vars.FromPtr(&binding),
		"abbr":    vars.FromPtr(&abbr),
		"quote-paste": vars.FromPtrWithMutex(
			&m.Config.Raw.QuotePaste, &m.Config.Mutex),
	}.AddBuiltinFns("[insert mode]", map[string]interface{}{
		"start": func() { ed.State().SetMode(m) },
		"default": func() error {
			return utils.ActionError(utils.BasicHandler(
				tty.KeyEvent(ed.State().BindingKey()), ed.State()))
		},
	})

	return m, ns
}

// Iterates through each pair in the hashmap and calls the callback for each
// pair, ignoring pairs that are mistyped.
func abbrIterate(abbr hashmap.Map, cb func(a, f string)) {
	for it := abbr.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		abbr, kOk := k.(string)
		full, vOk := v.(string)
		if !kOk || !vOk {
			continue
		}
		cb(abbr, full)
	}
}

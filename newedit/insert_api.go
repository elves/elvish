package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/insert"
	"github.com/xiaq/persistent/hashmap"
)

// Initializes states for the insert mode and its API.
func initInsert(nt notifier, ev *eval.Evaler) (*insert.Mode, eval.Ns) {
	abbr := vals.EmptyMap
	binding := EmptyBindingMap

	m := &insert.Mode{
		KeyHandler:  keyHandlerFromBinding(nt, ev, &binding),
		AbbrIterate: func(cb func(a, f string)) { abbrIterate(abbr, cb) },
	}

	ns := eval.NewNs().
		Add("binding", vars.FromPtr(&binding)).
		Add("abbr", vars.FromPtr(&abbr)).
		Add("quote-paste",
			vars.FromPtrWithMutex(&m.Config.Raw.QuotePaste, &m.Config.Mutex))

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

package eval

import "github.com/elves/elvish/store"

type sharedVariable struct {
	store *store.Store
	name  string
}

func (sv sharedVariable) Set(val Value) {
	err := sv.store.SetSharedVar(sv.name, ToString(val))
	maybeThrow(err)
}

func (sv sharedVariable) Get() Value {
	value, err := sv.store.GetSharedVar(sv.name)
	maybeThrow(err)
	return String(value)
}

package eval

import "github.com/elves/elvish/daemon"

type sharedVariable struct {
	store *daemon.Client
	name  string
}

func (sv sharedVariable) Set(val Value) {
	err := sv.store.SetSharedVar(sv.name, ToString(val))
	maybeThrow(err)
}

func (sv sharedVariable) Get() Value {
	value, err := sv.store.SharedVar(sv.name)
	maybeThrow(err)
	return String(value)
}

package etk

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
)

// Bespoke method map bindings for Etk types.
//
// In future, this file should be replaced with Elvish's general method map mechanism.

func (c Context) Fields() vals.PropertyMap { return contextFields{c} }

type contextFields struct{ c Context }

func (cf contextFields) AddMsg() eval.Callable   { return eval.NewGoFn("add-msg", cf.c.AddMsg) }
func (cf contextFields) Refresh() eval.Callable  { return eval.NewGoFn("refresh", cf.c.Refresh) }
func (cf contextFields) Finished() eval.Callable { return eval.NewGoFn("finished", cf.c.Finished) }
func (cf contextFields) Subcomp() eval.Callable  { return eval.NewGoFn("subcomp", cf.c.Subcomp) }
func (cf contextFields) Binding() eval.Callable  { return eval.NewGoFn("binding", cf.c.Binding) }
func (cf contextFields) BindingNopDefault() eval.Callable {
	return eval.NewGoFn("binding-nop-default", cf.c.BindingNopDefault)
}
func (cf contextFields) Get() eval.Callable   { return eval.NewGoFn("get", cf.c.Get) }
func (cf contextFields) Set() eval.Callable   { return eval.NewGoFn("set", cf.c.Set) }
func (cf contextFields) Frame() eval.Callable { return eval.NewGoFn("frame", cf.c.Frame) }
func (cf contextFields) UpdateAsync() eval.Callable {
	return eval.NewGoFn("update-async", cf.c.UpdateAsync)
}
func (cf contextFields) FinishChan() eval.Callable {
	return eval.NewGoFn("finish-chan", cf.c.FinishChan)
}

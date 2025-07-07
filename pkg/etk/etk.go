//go:generate stringer -type=Reaction -output=zstring.go
package etk

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/ui"
)

// TODO: Automatically remove state that hasn't been referenced?

// Comp is a component. It is called every time the state changes to generate a
// new pair of [View] and [React].
type Comp = func(Context) (View, React)

// WithBefore returns a variation of the [Comp] that runs a function every time
// before it is called.
func WithBefore(f Comp, before func(Context)) Comp {
	return func(c Context) (View, React) {
		before(c)
		return f(c)
	}
}

// WithInit returns a variation of the [Comp] that overrides the initial value
// of some state variables. The variadic arguments must come in (key, value)
// pairs, and the keys must be strings.
func WithInit(f Comp, pairs ...any) Comp {
	return func(c Context) (View, React) {
		for i := 0; i < len(pairs); i += 2 {
			key, value := pairs[i].(string), pairs[i+1]
			State(c, key, value)
		}
		return f(c)
	}
}

// React is how a component handles an event.
type React = func(term.Event) Reaction

// Reaction is the return value from [React].
type Reaction uint32

// Possible [Reaction] values.
const (
	Unused Reaction = iota
	Consumed
	Finish
	FinishEOF
)

type globalContext struct {
	// Used in StateVar, needed for converting Elvish functions to Go functions.
	fm *eval.Frame
	// Global states.
	state vals.Map
	msgs  []ui.Text
	// Access to global states is guarded by two mutexes:
	//
	// - Each individual state mutation is guarded by stateMutex automatically
	//   (implemented by StateVar).
	//
	// - Each "batch" of access is additionally guarded by batchMutex. A "batch"
	//   is a refresh, an event handling process, or an implicitly initiated
	//   batch. Use of batch mutex relies on co-operation.
	stateMutex sync.RWMutex
	batchMutex sync.Mutex
	// A channel for components to request a refresh, typically as a result of
	// some external asynchronous event. Must be a buffered channel.
	refreshCh chan struct{}
	// A channel that is closed when the event loop finishes. Goroutines spawned
	// by components should listen on this channel and terminate when it closes.
	finishCh chan struct{}
}

func makeGlobalContext(fm *eval.Frame) *globalContext {
	return &globalContext{
		state: vals.EmptyMap, fm: fm,
		refreshCh: make(chan struct{}, 1),
		finishCh:  make(chan struct{}),
	}
}

func (g *globalContext) PopMsgs() []ui.Text {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	msgs := g.msgs
	g.msgs = nil
	return msgs
}

// Context provides two kinds of context:
//
//   - The state subtree,
//     used for storing all the persistent state of a component.
//
//     The state subtree is just an Elvish map on the low level,
//     but Etk components should use [StateVar] to access it.
//     (The functions returning StateVar's are ideally methods of Context,
//     but since they have type parameters,
//     they have to be free functions rather than methods.)
//
//   - Some global ephemeral states and coordination mechanisms,
//     like the message buffer and channels for the rendering lifecycle.
type Context struct {
	g    *globalContext
	path []string
}

func (c Context) Frame() *eval.Frame { return c.g.fm }

func (c Context) UpdateAsync(f func()) {
	func() {
		c.g.batchMutex.Lock()
		defer c.g.batchMutex.Unlock()
		f()
	}()
	c.Refresh()
}

func (c Context) descPath(path ...string) []string {
	return slices.Concat(c.path, path)
}

// AddMsg adds a new message to the message buffer.
// When rendering finishes,
// all the messages in the buffer will get shown to the user,
// and the buffer gets cleared.
//
// This method always triggers a refresh.
// TODO: Don't refresh if we're currently in a rendering cycle.
func (c Context) AddMsg(msg ui.Text) {
	c.g.stateMutex.Lock()
	defer c.g.stateMutex.Unlock()
	c.g.msgs = append(c.g.msgs, msg)
	c.Refresh()
}

// Context requests a re-render.
// This is typically useful from asynchronous tasks.
func (c Context) Refresh() {
	select {
	case c.g.refreshCh <- struct{}{}:
	default:
	}
}

// FinishChan returns a channel that is closed when the event loop finishes.
//
// This is useful for asynchronous tasks:
// Goroutines spawned by components should listen on this channel,
// and terminate when it is closed.
// ([Context.Finished] provides an alternative.)
func (c Context) FinishChan() <-chan struct{} {
	return c.g.finishCh
}

// Finished returns whether the event loop has finished.
//
// This is an alternative to [Context.FinishChan]:
// If it's impractical for a goroutine to listen to FinishChan,
// it can call this function regularly,
// and terminate when it returns true.
func (c Context) Finished() bool {
	select {
	case <-c.g.finishCh:
		return true
	default:
		return false
	}
}

// Subcomp creates two state variables:
//
//   - A map, with the given name
//
//   - A Comp, with the given name plus "-comp" as its name and f as the initial
//     value
//
// It then invokes the Comp with the map as the context.
func (c Context) Subcomp(name string, f Comp) (View, React) {
	State(c, name, vals.EmptyMap)
	compVar := State(c, name+"-comp", f)
	return compVar.Get()(Context{c.g, c.descPath(name)})
}

type Binding = func(c Context, e term.Event) Reaction

func (c Context) Binding(def React) React {
	binding := State(c, "binding", Binding(nil)).Get()
	return func(e term.Event) Reaction {
		if binding != nil {
			return binding(c, e)
		} else {
			return def(e)
		}
	}
}

func (c Context) BindingNopDefault() React {
	return c.Binding(func(term.Event) Reaction { return Unused })
}

func (c Context) Get(key string) any {
	return BindState(c, key, any(nil)).getAny()
}

func (c Context) Set(key string, value any) {
	BindState(c, key, any(nil)).setAny(value)
}

// State returns a state variable with the given path from the current level,
// initializing it to a given value if it doesn't exist yet.
func State[T any](c Context, key string, initial T) StateVar[T] {
	sv := BindState(c, key, initial)
	if sv.getAny() == nil {
		sv.Set(initial)
	}
	return sv
}

// BindState returns a state variable with the given path from the current
// level. It doesn't initialize the variable.
//
// This should only be used if the variable is initialized elsewhere, most
// typically for accessing the state of a subcomponent after the subcomponent
// has been called.
func BindState[T any](c Context, key string, fallback T) StateVar[T] {
	path := c.descPath(strings.Split(key, "/")...)
	return StateVar[T]{&c.g.state, &c.g.stateMutex, c.g.fm, path, fallback}
}

// StateVar provides access to a "state variable",
// which is just a fancy name for an entry in the state map.
type StateVar[T any] struct {
	state    *vals.Map
	mutex    *sync.RWMutex
	fm       *eval.Frame
	path     []string
	fallback T
}

// TODO: Make access concurrency-correct with a pair of mutexes and an epoch

func (sv StateVar[T]) Get() T {
	raw := sv.getAny()
	val, err := ScanToGo[T](raw, sv.fm)
	if err == nil {
		return val
	}
	// TODO: Report the error somewhere?
	return sv.fallback
}

func (sv StateVar[T]) Set(t T) {
	sv.setAny(t)
}

func (sv StateVar[T]) Swap(f func(T) T) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()

	raw := sv.get()
	val, err := ScanToGo[T](raw, sv.fm)
	if err != nil {
		val = sv.fallback
	}
	sv.set(f(val))
}

// A variant of vals.ScanToGo, with additional support for adapting an Elvish
// function to a Go function.
func ScanToGo[T any](val any, fm *eval.Frame) (T, error) {
	var dst T
	err := vals.ScanToGo(val, &dst)
	if err == nil {
		return dst, nil
	}
	dstType := reflect.TypeFor[T]()
	if fn, ok := val.(eval.Callable); ok && dstType.Kind() == reflect.Func {
		// Adapt an Elvish function to a Go function
		return reflect.MakeFunc(dstType, func(args []reflect.Value) []reflect.Value {
			// TODO: Handle errors properly
			// TODO: Add intermediate "internal" entry to the traceback
			outs := must.OK1(fm.CaptureOutput(func(fm *eval.Frame) error {
				return fn.Call(fm, each(args, reflect.Value.Interface), eval.NoOpts)
			}))
			goOuts := make([]reflect.Value, dstType.NumOut())
			if len(outs) != len(goOuts) {
				panic("wrong number of outputs")
			}
			for i, out := range outs {
				goOutPtr := reflect.New(dstType.Out(i))
				must.OK(vals.ScanToGo(out, goOutPtr.Interface()))
				goOuts[i] = reflect.Indirect(goOutPtr)
			}
			return goOuts
		}).Interface().(T), nil
	}
	return zero[T](), err
}

func (sv StateVar[T]) getAny() any {
	sv.mutex.RLock()
	defer sv.mutex.RUnlock()
	return sv.get()
}

func (sv StateVar[T]) setAny(v any) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	sv.set(v)
}

func (sv StateVar[T]) get() any  { return getPath(*sv.state, sv.path) }
func (sv StateVar[T]) set(v any) { *sv.state = assocPath(*sv.state, sv.path, v) }

type StateSubTreeVar Context

func (v StateSubTreeVar) Get() any {
	return getPath(v.g.state, v.path)
}

func (v StateSubTreeVar) Set(val any) error {
	valMap, ok := val.(vals.Map)
	if !ok {
		return fmt.Errorf("must be map")
	}
	v.g.state = assocPath(v.g.state, v.path, valMap)
	return nil
}

// TODO: Move the following to vals?

func getPath[T any](m vals.Map, path []T) any {
	if len(path) == 0 {
		return m
	}
	for len(path) > 1 {
		v, _ := m.Index(path[0])
		if v == nil {
			return nil
		}
		m = v.(vals.Map)
		path = path[1:]
	}
	v, _ := m.Index(path[0])
	return v
}

func assocPath[T any](m vals.Map, path []T, newVal any) vals.Map {
	if len(path) == 0 {
		return newVal.(vals.Map)
	}

	if len(path) == 1 {
		return m.Assoc(path[0], newVal)
	}
	v, _ := m.Index(path[0])
	if v == nil {
		v = vals.EmptyMap
	}
	return m.Assoc(path[0], assocPath(v.(vals.Map), path[1:], newVal))
}

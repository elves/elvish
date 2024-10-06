// Package etk implements an [immediate mode] TUI framework with managed states.
//
// Each component in the TUI is implemented by a [Comp]: a function taking a
// [Context] and returning a [View] and a [React]:
//
//   - The [Context] provides access to states associated with the component and
//     supports creating sub-components.
//
//   - The [View] is a snapshot of the UI that reflects the current state.
//
//   - The [React] is a function that will be called to react to an event.
//
// Whenever there is an update in the states, the function is called again to
// generate a pair of [View] and [React].
//
// The state is organized into a tree, with individual state variables as leaf
// nodes and components as inner nodes. The [Context] provides access to the
// current level and all descendant levels, allowing a component to manipulate
// not just its own state, but also that of any descendant. This is the only way
// of passing information between components: if a component has any
// customizable property, it is modelled as a state that its parent can modify.
//
// # Design notes
//
// Immediate mode is an alternative to the more common [retained mode] style of
// graphics API. Some GUI frameworks using this style are [Dear ImGui] and [Gio
// UI]. [React], [SwiftUI] and [Jetpack Compose] also provide immediate mode
// APIs above an underlying [retained mode] API.
//
// Immediate mode libraries differ a lot in how component structure and state
// are managed. Etk is used to implement Elvish's terminal UI, so the choices
// made by etk is driven largely by how easy it is to create an Elvish binding
// for the framework that is maximally programmable:
//
//   - The open nature of the state tree makes it easy to inspect and mutate the
//     terminal UI as it is running.
//
//   - The managed nature of the state tree gives us concurrency safety and
//     undo/redo almost for free.
//
//   - The use of [vals.Map] to back the state tree sacrifices type safety in
//     the Go version of the framework, but makes Elvish integration much
//     easier.
//
// [immediate mode]: https://en.wikipedia.org/wiki/Immediate_mode_(computer_graphics)
// [retained mode]: https://en.wikipedia.org/wiki/Retained_mode
// [Dear ImGui]: https://github.com/ocornut/imgui
// [Gio UI]: https://gioui.org
// [React]: https://react.dev
// [SwiftUI]: https://developer.apple.com/xcode/swiftui/
// [Jetpack Compose]: https://developer.android.com/compose
//
//go:generate stringer -type=Reaction -output=zstring.go
package etk

import (
	"io"
	"reflect"
	"slices"
	"strings"

	"src.elv.sh/pkg/cli"
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
	state     vals.Map
	refreshCh chan struct{}
	finishCh  chan struct{}
	fm        *eval.Frame
	msgs      []ui.Text
}

// Context provides access to the state tree at the current level and all
// descendant levels.
type Context struct {
	g    *globalContext
	path []string
}

func (c Context) Frame() *eval.Frame { return c.g.fm }

func (c Context) descPath(path ...string) []string {
	return slices.Concat(c.path, path)
}

func (c Context) Notify(msg ui.Text) {
	// TODO: concurrency-safety
	c.g.msgs = append(c.g.msgs, msg)
}

func (c Context) Refresh() {
	select {
	case c.g.refreshCh <- struct{}{}:
	default:
	}
}

func (c Context) FinishChan() <-chan struct{} {
	return c.g.finishCh
}

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

// Binding customizes how a component reacts to an event.
type Binding func(ev term.Event, c Context, r React) Reaction

// WithBinding wraps a [React] function, adding support for using the "binding"
// state variable as the [Binding].
func (c Context) WithBinding(f React) React {
	bindingVar := State(c, "binding", Binding(nil))
	return func(ev term.Event) Reaction {
		if binding := bindingVar.Get(); binding != nil {
			return binding(ev, c, f)
		}
		return f(ev)
	}
}

func (c Context) Get(key string) any {
	return getPath(c.g.state, c.descPath(strings.Split(key, "/")...))
}

func (c Context) Set(key string, value any) {
	c.g.state = assocPath(c.g.state, c.descPath(strings.Split(key, "/")...), value)
}

// State returns a state variable with the given name at the current level,
// initializing it to a given value if it doesn't exist yet.
func State[T any](c Context, name string, initial T) StateVar[T] {
	sv := BindState[T](c, name, initial)
	if sv.getRaw() == nil {
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
func BindState[T any](c Context, name string, fallback T) StateVar[T] {
	path := c.descPath(strings.Split(name, "/")...)
	return StateVar[T]{&c.g.state, c.g.fm, path, fallback}
}

// StateVar provides access to a state variable, a node in the state tree.
type StateVar[T any] struct {
	state    *vals.Map
	fm       *eval.Frame
	path     []string
	fallback T
}

// TODO: Make access concurrency-correct with a pair of mutexes and an epoch

func (sv StateVar[T]) Get() T {
	raw := sv.getRaw()
	val, err := ScanToGo[T](raw, sv.fm)
	if err == nil {
		return val
	}
	// TODO: Report the error somewhere?
	return sv.fallback
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

func (sv StateVar[T]) getRaw() any { return getPath(*sv.state, sv.path) }

func (sv StateVar[T]) Set(t T)          { *sv.state = assocPath(*sv.state, sv.path, t) }
func (sv StateVar[T]) Swap(f func(T) T) { sv.Set(f(sv.Get())) }

func getPath(m vals.Map, path []string) any {
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

func assocPath(m vals.Map, path []string, newVal any) vals.Map {
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

func Run(tty cli.TTY, fm *eval.Frame, f Comp) (vals.Map, error) {
	restore, err := tty.Setup()
	if err != nil {
		return nil, err
	}
	defer restore()

	// Start reading events.
	eventCh := make(chan term.Event)
	go func() {
		for {
			event, err := tty.ReadEvent()
			if err != nil {
				if err == term.ErrStopped {
					return
				}
				// TODO: Report error in notification
			}
			eventCh <- event
		}
	}()
	defer tty.CloseReader()

	sc := Stateful(fm, f)
	defer sc.Finish()

	for {
		// Render.
		h, w := tty.Size()
		buf := sc.Render(w, h)
		msgBuf := sc.RenderAndPopMsgs(w)
		tty.UpdateBuffer(msgBuf, buf, false /*true*/)

		select {
		case event := <-eventCh:
			reaction := sc.Handle(event)
			if reaction == Finish || reaction == FinishEOF {
				h, w := tty.Size()
				buf := sc.Render(w, h)
				msgBuf := sc.RenderAndPopMsgs(w)
				// Render the final view with a trailing newline. This operation
				// is quite subtle with the term.Buffer API.
				buf.Extend(term.NewBufferBuilder(w).Buffer(), true)
				tty.UpdateBuffer(msgBuf, buf, false)
				if reaction == FinishEOF {
					return sc.g.state, io.EOF
				} else {
					return sc.g.state, nil
				}
			}
		case <-sc.g.refreshCh:
			sc.Refresh()
		}
	}
}

package newedit

import (
	"bufio"
	"io"
	"os"
	"sync"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/newedit/editutil"
	"github.com/elves/elvish/newedit/types"
)

// TODO(xiaq): Move the implementation into this package.

// BindingMap is a specialized map type for key bindings.
type BindingMap = eddefs.BindingMap

// EmptyBindingMap is an empty binding map. It is useful for building binding
// maps.
var EmptyBindingMap = eddefs.EmptyBindingMap

func keyHandlerFromBinding(ed editor, ev *eval.Evaler, m *BindingMap) func(ui.Key) types.HandlerAction {
	return func(k ui.Key) types.HandlerAction {
		f := m.GetOrDefault(k)
		// TODO: Make this fallback part of GetOrDefault after moving BindingMap
		// into this package.
		if f == nil {
			ed.Notify("Unbound: " + k.String())
			return types.NoAction
		}
		ed.State().SetBindingKey(k)
		return callBinding(ed, ev, f)
	}
}

var bindingSource = eval.NewInternalSource("[editor binding]")

func callBinding(nt notifier, ev *eval.Evaler, f eval.Callable) types.HandlerAction {

	// TODO(xiaq): Use CallWithOutputCallback when it supports redirecting the
	// stderr port.
	notifyPort, cleanup := makeNotifyPort(nt.Notify)
	defer cleanup()
	ports := []*eval.Port{eval.DevNullClosedChan, notifyPort, notifyPort}
	frame := eval.NewTopFrame(ev, bindingSource, ports)

	err := frame.Call(f, nil, eval.NoOpts)

	if err != nil {
		if action, ok := eval.Cause(err).(editutil.ActionError); ok {
			return types.HandlerAction(action)
		}
		// TODO(xiaq): Make the stack trace available.
		nt.Notify("[binding error] " + err.Error())
	}
	return types.NoAction
}

func makeNotifyPort(notify func(string)) (*eval.Port, func()) {
	ch := make(chan interface{})
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		// Relay value outputs
		for v := range ch {
			notify("[value out] " + vals.Repr(v, vals.NoPretty))
		}
		wg.Done()
	}()
	go func() {
		// Relay byte outputs
		reader := bufio.NewReader(r)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if line != "" {
					notify("[bytes out] " + line)
				}
				if err != io.EOF {
					notify("[bytes error] " + err.Error())
				}
				break
			}
			notify("[bytes out] " + line[:len(line)-1])
		}
		wg.Done()
	}()
	port := &eval.Port{Chan: ch, File: w, CloseChan: true, CloseFile: true}
	cleanup := func() {
		port.Close()
		wg.Wait()
	}
	return port, cleanup
}

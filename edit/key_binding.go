package edit

import (
	"bufio"
	"io"
	"os"
	"sync"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/ui"
)

type mapBinding struct {
	nt      notifier
	ev      *eval.Evaler
	mapVars []vars.PtrVar
}

func newMapBinding(nt notifier, ev *eval.Evaler, mapVars ...vars.PtrVar) el.Handler {
	return mapBinding{nt, ev, mapVars}
}

func (b mapBinding) Handle(e term.Event) bool {
	k, ok := e.(term.KeyEvent)
	if !ok {
		return false
	}
	maps := make([]BindingMap, len(b.mapVars))
	for i, v := range b.mapVars {
		maps[i] = v.GetRaw().(BindingMap)
	}
	f := indexLayeredBindings(ui.Key(k), maps...)
	if f == nil {
		return false
	}
	callWithNotifyPorts(b.nt, b.ev, f)
	return true
}

// Indexes a series of layered bindings. Returns nil if none of the bindings
// have the required key or a default.
func indexLayeredBindings(k ui.Key, bindings ...BindingMap) eval.Callable {
	for _, binding := range bindings {
		if binding.HasKey(k) {
			return binding.GetKey(k)
		}
	}
	for _, binding := range bindings {
		if binding.HasKey(ui.Default) {
			return binding.GetKey(ui.Default)
		}
	}
	return nil
}

var bindingSource = eval.NewInternalSource("[editor binding]")

func callWithNotifyPorts(nt notifier, ev *eval.Evaler, f eval.Callable, args ...interface{}) {
	// TODO(xiaq): Use CallWithOutputCallback when it supports redirecting the
	// stderr port.
	notifyPort, cleanup := makeNotifyPort(nt.Notify)
	defer cleanup()
	ports := []*eval.Port{eval.DevNullClosedChan, notifyPort, notifyPort}
	frame := eval.NewTopFrame(ev, bindingSource, ports)

	err := frame.Call(f, args, eval.NoOpts)
	if err != nil {
		// TODO(xiaq): Make the stack trace available.
		nt.Notify("[binding error] " + err.Error())
	}
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

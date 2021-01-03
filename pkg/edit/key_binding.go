package edit

import (
	"bufio"
	"io"
	"os"
	"sync"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/ui"
)

type mapBinding struct {
	nt      notifier
	ev      *eval.Evaler
	mapVars []vars.PtrVar
}

func newMapBinding(nt notifier, ev *eval.Evaler, mapVars ...vars.PtrVar) cli.Handler {
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

var bindingSource = parse.Source{Name: "[editor binding]"}

func callWithNotifyPorts(nt notifier, ev *eval.Evaler, f eval.Callable, args ...interface{}) {
	notifyPort, cleanup := makeNotifyPort(nt)
	defer cleanup()

	err := ev.Call(f,
		eval.CallCfg{Args: args, From: "[editor binding]"},
		eval.EvalCfg{Ports: []*eval.Port{nil, notifyPort, notifyPort}})
	if err != nil {
		nt.notifyError("binding", err)
	}
}

func makeNotifyPort(nt notifier) (*eval.Port, func()) {
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
			nt.notifyf("[value out] %s", vals.Repr(v, vals.NoPretty))
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
					nt.notifyf("[bytes out] %s", line)
				}
				if err != io.EOF {
					nt.notifyf("[bytes error] %s", err)
				}
				break
			}
			nt.notifyf("[bytes out] %s", line[:len(line)-1])
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

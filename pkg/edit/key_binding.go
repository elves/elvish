package edit

import (
	"bufio"
	"io"
	"os"
	"sync"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/ui"
)

type mapBindings struct {
	nt      notifier
	ev      *eval.Evaler
	mapVars []vars.PtrVar
}

func newMapBindings(nt notifier, ev *eval.Evaler, mapVars ...vars.PtrVar) tk.Bindings {
	return mapBindings{nt, ev, mapVars}
}

func (b mapBindings) Handle(w tk.Widget, e term.Event) bool {
	k, ok := e.(term.KeyEvent)
	if !ok {
		return false
	}
	maps := make([]bindingsMap, len(b.mapVars))
	for i, v := range b.mapVars {
		maps[i] = v.GetRaw().(bindingsMap)
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
func indexLayeredBindings(k ui.Key, maps ...bindingsMap) eval.Callable {
	for _, m := range maps {
		if m.HasKey(k) {
			return m.GetKey(k)
		}
	}
	for _, m := range maps {
		if m.HasKey(ui.DefaultKey) {
			return m.GetKey(ui.DefaultKey)
		}
	}
	return nil
}

func callWithNotifyPorts(nt notifier, ev *eval.Evaler, f eval.Callable, args ...any) {
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
	ch := make(chan any)
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		// Relay value outputs
		for v := range ch {
			nt.notifyf("[value out] %s", vals.ReprPlain(v))
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
		r.Close()
		wg.Done()
	}()
	port := &eval.Port{Chan: ch, File: w}
	cleanup := func() {
		close(ch)
		w.Close()
		wg.Wait()
	}
	return port, cleanup
}

package edit

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/ui"
)

func startInstant(c etk.Context) {
}

func startCommand(c etk.Context) {
}

func pendingVar(c etk.Context) etk.StateVar[comps.PendingText] {
	return etk.BindState(c, "code/pending", comps.PendingText{})
}

func etkBindingFromBindingMap(ed *Editor, m *bindingsMap) etk.Binding {
	return func(c etk.Context, ev term.Event) etk.Reaction {
		handled := ed.callBinding(m, ev)
		if handled {
			return etk.Consumed
		} else {
			return etk.Unused
		}
	}
}

// Duplicate with pkg/edit/key_binding.go
func makeNotifyPort(c etk.Context) (*eval.Port, func()) {
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
			notifyf(c, "[value out] %s", vals.ReprPlain(v))
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
					notifyf(c, "[bytes out] %s", line)
				}
				if err != io.EOF {
					notifyf(c, "[bytes error] %s", err)
				}
				break
			}
			notifyf(c, "[bytes out] %s", line[:len(line)-1])
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

func notifyf(c etk.Context, format string, args ...any) {
	c.AddMsg(ui.T(fmt.Sprintf(format, args...)))
}

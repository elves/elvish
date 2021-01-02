package edit

import (
	"bufio"
	"io"
	"os"
	"sync"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/addons/instant"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/strutil"
)

//elvdoc:var -instant:binding
//
// Binding for the instant mode.

//elvdoc:fn -instant:start
//
// Starts the instant mode. In instant mode, any text entered at the command
// line is evaluated immediately, with the output displayed.
//
// **WARNING**: Beware of unintended consequences when using destructive
// commands. For example, if you type `sudo rm -rf /tmp/*` in the instant mode,
// Elvish will attempt to evaluate `sudo rm -rf /` when you typed that far.

func initInstant(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(ed, ev, bindingVar)
	nb.AddNs("-instant",
		eval.NsBuilder{
			"binding": bindingVar,
		}.AddGoFns("<edit:-instant>:", map[string]interface{}{
			"start": func() { instantStart(ed.app, ev, binding) },
		}).Ns())
}

func instantStart(app cli.App, ev *eval.Evaler, binding cli.Handler) {
	execute := func(code string) ([]string, error) {
		outPort, collect, err := capturePort()
		if err != nil {
			return nil, err
		}
		err = ev.Eval(
			parse.Source{Name: "[instant]", Code: code},
			eval.EvalCfg{
				Ports:     []*eval.Port{nil, outPort},
				Interrupt: eval.ListenInterrupts})
		return collect(), err
	}
	instant.Start(app, instant.Config{Binding: binding, Execute: execute})
}

// Like eval.CapturePort, but collects output into strings, with value outputs
// stringified and prepended a marker.
func capturePort() (*eval.Port, func() []string, error) {
	var lines []string
	var mu sync.Mutex
	addLine := func(line string) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, line)
	}
	port, done, err := eval.PipePort(
		func(ch <-chan interface{}) {
			for v := range ch {
				addLine("▶ " + vals.ToString(v))
			}
		},
		func(r *os.File) {
			bufr := bufio.NewReader(r)
			for {
				line, err := bufr.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						addLine("i/o error: " + err.Error())
					}
					break
				}
				addLine(strutil.ChopLineEnding(line))
			}
		})
	if err != nil {
		return nil, nil, err
	}
	return port, func() []string {
		done()
		return lines
	}, nil
}

package edit

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/addons/instant"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
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

func initInstant(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddNs("-instant",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:-instant>:", map[string]interface{}{
			"start": func() { instantStart(app, ev, binding) },
		}))
}

func instantStart(app cli.App, ev *eval.Evaler, binding cli.Handler) {
	execute := func(code string) ([]string, error) {
		src := eval.NewInteractiveSource(code)
		n, err := parse.AsChunk("[instant]", code)
		if err != nil {
			return nil, err
		}
		op, err := ev.Compile(n, src)
		if err != nil {
			return nil, err
		}
		fm := eval.NewTopFrame(ev, src, []*eval.Port{
			{File: eval.DevNull},
			{}, // Will be replaced in CaptureOutput
			{File: eval.DevNull},
		})
		var output []string
		var outputMutex sync.Mutex
		addLine := func(line string) {
			outputMutex.Lock()
			defer outputMutex.Unlock()
			output = append(output, line)
		}
		valuesCb := func(ch <-chan interface{}) {
			for v := range ch {
				addLine("▶ " + vals.ToString(v))
			}
		}
		bytesCb := func(r *os.File) {
			bufr := bufio.NewReader(r)
			for {
				line, err := bufr.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						addLine("i/o error: " + err.Error())
					}
					break
				}
				addLine(strings.TrimSuffix(line, "\n"))
			}
		}
		err = fm.PipeOutput(
			func(fm *eval.Frame) error { return fm.Eval(op) }, valuesCb, bytesCb)
		return output, err
	}
	instant.Start(app, instant.Config{Binding: binding, Execute: execute})
}

package edit

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/addons/instant"
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
)

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

func instantStart(app cli.App, ev *eval.Evaler, binding el.Handler) {
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
		err = fm.ExecWithOutputCallback(op, valuesCb, bytesCb)
		return output, err
	}
	instant.Start(app, instant.Config{Binding: binding, Execute: execute})
}

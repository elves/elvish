package edit

import (
	"bufio"
	"os"
	"sync"

	"github.com/elves/elvish/eval"
)

// CallFn calls an Fn, displaying its outputs and possible errors as editor
// notifications. It is the preferred way to call a Fn while the editor is
// active.
func (ed *Editor) CallFn(fn eval.FnValue, args ...eval.Value) {
	if b, ok := fn.(*BuiltinFn); ok {
		// Builtin function: quick path.
		b.impl(ed)
		return
	}

	rout, chanOut, ports, err := makePorts()
	if err != nil {
		return
	}

	// Goroutines to collect output.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		rd := bufio.NewReader(rout)
		for {
			line, err := rd.ReadString('\n')
			if err != nil {
				break
			}
			ed.Notify("[bytes output] %s", line[:len(line)-1])
		}
		rout.Close()
		wg.Done()
	}()
	go func() {
		for v := range chanOut {
			ed.Notify("[value output] %s", v.Repr(eval.NoPretty))
		}
		wg.Done()
	}()

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor]", "", ports)
	ex := ec.PCall(fn, args, eval.NoOpts)
	if ex != nil {
		ed.Notify("function error: %s", ex.Error())
	}

	eval.ClosePorts(ports)
	wg.Wait()
	ed.refresh(true, true)
}

// makePorts connects stdin to /dev/null and a closed channel, identifies
// stdout and stderr and connects them to a pipe and channel. It returns the
// other end of stdout and the resulting []*eval.Port. The caller is
// responsible for closing the returned file and calling eval.ClosePorts on the
// ports.
func makePorts() (*os.File, chan eval.Value, []*eval.Port, error) {
	in, err := makeClosedStdin()
	if err != nil {
		return nil, nil, nil, err
	}

	// Output
	rout, out, err := os.Pipe()
	if err != nil {
		Logger.Println(err)
		return nil, nil, nil, err
	}
	chanOut := make(chan eval.Value)

	return rout, chanOut, []*eval.Port{
		in,
		{File: out, CloseFile: true, Chan: chanOut, CloseChan: true},
		{File: out, Chan: chanOut},
	}, nil
}

func makeClosedStdin() (*eval.Port, error) {
	// Input
	devnull, err := os.Open("/dev/null")
	if err != nil {
		Logger.Println(err)
		return nil, err
	}
	in := make(chan eval.Value)
	close(in)
	return &eval.Port{File: devnull, CloseFile: true, Chan: in}, nil
}

package edit

import (
	"bufio"
	"errors"
	"os"
	"sync"

	"github.com/elves/elvish/eval"
)

var (
	DevNull         *os.File
	ClosedChan      chan eval.Value
	NullClosedInput *eval.Port
)

func init() {
	var err error
	DevNull, err = os.Open("/dev/null")
	if err != nil {
		os.Stderr.WriteString("cannot open /dev/null, shell might not function normally")
	}
	ClosedChan = make(chan eval.Value)
	close(ClosedChan)
	NullClosedInput = &eval.Port{File: DevNull, Chan: ClosedChan}
}

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
	// Output
	rout, out, err := os.Pipe()
	if err != nil {
		Logger.Println(err)
		return nil, nil, nil, err
	}
	chanOut := make(chan eval.Value)

	return rout, chanOut, []*eval.Port{
		NullClosedInput,
		{File: out, CloseFile: true, Chan: chanOut, CloseChan: true},
		{File: out, Chan: chanOut},
	}, nil
}

// callFnAsPrompt calls a Fn with closed input, captures its output and convert
// the output to a slice of *styled's.
func callFnForPrompt(ed *Editor, fn eval.Fn) []*styled {
	ports := []*eval.Port{NullClosedInput, &eval.Port{File: os.Stdout}, &eval.Port{File: os.Stderr}}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor prompt]", "", ports)
	values, err := ec.PCaptureOutput(fn, nil, eval.NoOpts)
	if err != nil {
		ed.Notify("prompt function error: %v", err)
		return nil
	}

	var ss []*styled
	for _, v := range values {
		if s, ok := v.(*styled); ok {
			ss = append(ss, s)
		} else {
			ss = append(ss, &styled{eval.ToString(v), ""})
		}
	}
	return ss
}

func callFnForCandidates(fn eval.FnValue, ev *eval.Evaler, args []string) ([]*candidate, error) {
	ports := []*eval.Port{NullClosedInput, &eval.Port{File: os.Stdout}, &eval.Port{File: os.Stderr}}

	argValues := make([]eval.Value, len(args))
	for i, arg := range args {
		argValues[i] = eval.String(arg)
	}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ev, "[editor completer]", "", ports)
	values, err := ec.PCaptureOutput(fn, argValues, eval.NoOpts)
	if err != nil {
		return nil, errors.New("completer error: " + err.Error())
	}

	cands := make([]*candidate, len(values))
	for i, v := range values {
		switch v := v.(type) {
		case eval.String:
			cands[i] = &candidate{text: string(v)}
		case *candidate:
			cands[i] = v
		default:
			return nil, errors.New("completer must output string or candidate")
		}
	}
	return cands, nil
}

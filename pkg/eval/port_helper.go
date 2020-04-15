package eval

import (
	"os"
	"sync"

	"github.com/elves/elvish/pkg/eval/vals"
)

const (
	stdoutChanSize = 32
	stderrChanSize = 32
)

// DevNullPorts is 3 placeholder ports.
var DevNullPorts = [3]*Port{
	DevNullClosedChan, DevNullBlackholeChan, DevNullBlackholeChan}

// PortsFromFiles builds 3 ports from 3 files. It also returns a function that
// should be called when the ports are no longer needed.
func PortsFromFiles(files [3]*os.File, ev *Evaler) ([3]*Port, func()) {
	return portsFromFiles(files, ev.state.getValuePrefix())
}

func portsFromFiles(files [3]*os.File, prefix string) ([3]*Port, func()) {
	stdoutChan := make(chan interface{}, stdoutChanSize)
	stderrChan := make(chan interface{}, stderrChanSize)

	relayerWait := new(sync.WaitGroup)
	relayerWait.Add(2)
	go relayChanToFile(stdoutChan, files[1], prefix, relayerWait)
	go relayChanToFile(stderrChan, files[2], prefix, relayerWait)
	ports := [3]*Port{
		{File: files[0], Chan: ClosedChan},
		{File: files[1], Chan: stdoutChan, CloseChan: true},
		{File: files[2], Chan: stderrChan, CloseChan: true},
	}

	return ports, func() {
		close(stdoutChan)
		close(stderrChan)
		relayerWait.Wait()
	}
}

func relayChanToFile(ch <-chan interface{}, file *os.File, prefix string, w *sync.WaitGroup) {
	for v := range ch {
		file.WriteString(prefix)
		file.WriteString(vals.Repr(v, initIndent))
		file.WriteString("\n")
	}
	w.Done()
}

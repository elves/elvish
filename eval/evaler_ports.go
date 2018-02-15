package eval

import (
	"os"
	"sync"

	"github.com/elves/elvish/eval/vals"
)

const (
	stdoutChanSize = 32
	stderrChanSize = 32
)

type evalerPorts struct {
	ports      [3]*Port
	relayeWait *sync.WaitGroup
}

func newEvalerPorts(stdin, stdout, stderr *os.File, prefix *string) evalerPorts {
	stdoutChan := make(chan interface{}, stdoutChanSize)
	stderrChan := make(chan interface{}, stderrChanSize)

	var relayerWait sync.WaitGroup
	relayerWait.Add(2)
	go relayChanToFile(stdoutChan, stdout, prefix, &relayerWait)
	go relayChanToFile(stderrChan, stderr, prefix, &relayerWait)

	return evalerPorts{
		[3]*Port{
			{File: stdin, Chan: ClosedChan},
			{File: stdout, Chan: stdoutChan, CloseChan: true},
			{File: stderr, Chan: stderrChan, CloseChan: true},
		},
		&relayerWait,
	}
}

func relayChanToFile(ch <-chan interface{}, file *os.File, prefix *string, w *sync.WaitGroup) {
	for v := range ch {
		file.WriteString(*prefix)
		file.WriteString(vals.Repr(v, initIndent))
		file.WriteString("\n")
	}
	w.Done()
}

func (ep *evalerPorts) close() {
	ep.ports[1].Close()
	ep.ports[2].Close()
	ep.relayeWait.Wait()
}

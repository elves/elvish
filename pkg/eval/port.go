package eval

import (
	"fmt"
	"os"
)

// Port conveys data stream. It always consists of a byte band and a channel band.
type Port struct {
	File      *os.File
	Chan      chan interface{}
	CloseFile bool
	CloseChan bool
}

// Fork returns a copy of a Port with the Close* flags unset.
func (p *Port) Fork() *Port {
	return &Port{p.File, p.Chan, false, false}
}

// Close closes a Port.
func (p *Port) Close() {
	if p == nil {
		return
	}
	if p.CloseFile {
		p.File.Close()
	}
	if p.CloseChan {
		close(p.Chan)
	}
}

var (
	// ClosedChan is a closed channel, suitable for use as placeholder channel input.
	ClosedChan = getClosedChan()
	// BlackholeChan is channel writes onto which disappear, suitable for use as
	// placeholder channel output.
	BlackholeChan = getBlackholeChan()
	// DevNull is /dev/null.
	DevNull = getDevNull()
	// DevNullClosedChan is a port made up from DevNull and ClosedChan,
	// suitable as placeholder input port.
	DevNullClosedChan = &Port{File: DevNull, Chan: ClosedChan}
	// DevNullBlackholeChan is a port made up from DevNull and BlackholeChan,
	// suitable as placeholder output port.
	DevNullBlackholeChan = &Port{File: DevNull, Chan: BlackholeChan}
)

func getClosedChan() chan interface{} {
	ch := make(chan interface{})
	close(ch)
	return ch
}

func getBlackholeChan() chan interface{} {
	ch := make(chan interface{})
	go func() {
		for range ch {
		}
	}()
	return ch
}

func getDevNull() *os.File {
	f, err := os.Open(os.DevNull)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"cannot open %s, shell might not function normally\n", os.DevNull)
	}
	return f
}

package eval

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/strutil"
)

// Port conveys data stream. It always consists of a byte band and a channel band.
type Port struct {
	File      *os.File
	Chan      chan interface{}
	closeFile bool
	closeChan bool
}

// Returns a copy of the Port with the Close* flags unset.
func (p *Port) fork() *Port {
	return &Port{p.File, p.Chan, false, false}
}

// Closes a Port.
func (p *Port) close() {
	if p == nil {
		return
	}
	if p.closeFile {
		p.File.Close()
	}
	if p.closeChan {
		close(p.Chan)
	}
}

var (
	// ClosedChan is a closed channel, suitable as a placeholder input channel.
	ClosedChan = getClosedChan()
	// BlackholeChan is a channel that absorbs all values written to it,
	// suitable as a placeholder output channel.
	BlackholeChan = getBlackholeChan()
	// DevNull is /dev/null, suitable as a placeholder file for either input or
	// output.
	DevNull = getDevNull()

	// DummyInputPort is a port made up from DevNull and ClosedChan, suitable as
	// a placeholder input port.
	DummyInputPort = &Port{File: DevNull, Chan: ClosedChan}
	// DummyOutputPort is a port made up from DevNull and BlackholeChan,
	// suitable as a placeholder output port.
	DummyOutputPort = &Port{File: DevNull, Chan: BlackholeChan}
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

// PipePort returns an output *Port whose value and byte components are both
// piped. The supplied functions are called on a separate goroutine with the
// read ends of the value and byte components of the port. It also returns a
// function to clean up the port and wait for the callbacks to finish.
func PipePort(vCb func(<-chan interface{}), bCb func(*os.File)) (*Port, func(), error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	ch := make(chan interface{}, outputCaptureBufferSize)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		vCb(ch)
	}()
	go func() {
		defer wg.Done()
		defer r.Close()
		bCb(r)
	}()

	port := &Port{Chan: ch, closeChan: true, File: w, closeFile: true}
	done := func() {
		port.close()
		wg.Wait()
	}
	return port, done, nil
}

// CapturePort returns an output *Port whose value and byte components are
// both connected to an internal pipe that saves the output. It also returns a
// function to call to obtain the captured output.
func CapturePort() (*Port, func() []interface{}, error) {
	vs := []interface{}{}
	var m sync.Mutex
	port, done, err := PipePort(
		func(ch <-chan interface{}) {
			for v := range ch {
				m.Lock()
				vs = append(vs, v)
				m.Unlock()
			}
		},
		func(r *os.File) {
			buffered := bufio.NewReader(r)
			for {
				line, err := buffered.ReadString('\n')
				if line != "" {
					v := strutil.ChopLineEnding(line)
					m.Lock()
					vs = append(vs, v)
					m.Unlock()
				}
				if err != nil {
					if err != io.EOF {
						logger.Println("error on reading:", err)
					}
					break
				}
			}
		})
	if err != nil {
		return nil, nil, err
	}
	return port, func() []interface{} {
		done()
		return vs
	}, nil
}

// StringCapturePort is like CapturePort, but processes value outputs by
// stringifying them and prepending an output marker.
func StringCapturePort() (*Port, func() []string, error) {
	var lines []string
	var mu sync.Mutex
	addLine := func(line string) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, line)
	}
	port, done, err := PipePort(
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

// Buffer size for the channel to use in FilePort. The value has been chosen
// arbitrarily.
const filePortChanSize = 32

// FilePort returns an output *Port where the byte component is the file itself,
// and the value component is converted to an internal channel that writes
// each value to the file, prepending with a prefix. It also returns a cleanup
// function, which should be called when the *Port is no longer needed.
func FilePort(f *os.File, valuePrefix string) (*Port, func()) {
	ch := make(chan interface{}, filePortChanSize)
	relayDone := make(chan struct{})
	go func() {
		for v := range ch {
			f.WriteString(valuePrefix)
			f.WriteString(vals.Repr(v, vals.NoPretty))
			f.WriteString("\n")
		}
		close(relayDone)
	}()
	return &Port{File: f, Chan: ch}, func() {
		close(ch)
		<-relayDone
	}
}

// PortsFromStdFiles is a shorthand for calling PortsFromFiles with os.Stdin,
// os.Stdout and os.Stderr.
func PortsFromStdFiles(prefix string) ([]*Port, func()) {
	return PortsFromFiles([3]*os.File{os.Stdin, os.Stdout, os.Stderr}, prefix)
}

// PortsFromFiles builds 3 ports from 3 files. It also returns a function that
// should be called when the ports are no longer needed.
func PortsFromFiles(files [3]*os.File, prefix string) ([]*Port, func()) {
	port1, cleanup1 := FilePort(files[1], prefix)
	port2, cleanup2 := FilePort(files[2], prefix)
	return []*Port{{File: files[0], Chan: ClosedChan}, port1, port2}, func() {
		cleanup1()
		cleanup2()
	}
}

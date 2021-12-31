package eval

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/strutil"
)

// Port conveys data stream. It always consists of a byte band and a channel band.
type Port struct {
	File      *os.File
	Chan      chan interface{}
	closeFile bool
	closeChan bool

	// The following two fields are populated as an additional control
	// mechanism for output ports. When no more value should be send on Chan,
	// chanSendError is populated and chanSendStop is closed. This is used for
	// both detection of reader termination (see readerGone below) and closed
	// ports.
	sendStop  chan struct{}
	sendError *error

	// Only populated in output ports writing to another command in a pipeline.
	// When the reading end of the pipe exits, it stores 1 in readerGone. This
	// is used to check if an external command killed by SIGPIPE is caused by
	// the termination of the reader of the pipe.
	readerGone *int32
}

// ErrNoValueOutput is thrown when writing to a pipe without a value output
// component.
var ErrNoValueOutput = errors.New("port has no value output")

// A closed channel, suitable as a value for Port.sendStop when there is no
// reader to start with.
var closedSendStop = make(chan struct{})

func init() { close(closedSendStop) }

// Returns a copy of the Port with the Close* flags unset.
func (p *Port) fork() *Port {
	return &Port{p.File, p.Chan, false, false, p.sendStop, p.sendError, p.readerGone}
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

	// DummyPorts contains 3 dummy ports, suitable as stdin, stdout and stderr.
	DummyPorts = []*Port{DummyInputPort, DummyOutputPort, DummyOutputPort}
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
			f.WriteString(vals.ReprPlain(v))
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

// ValueOutput defines the interface through which builtin commands access the
// value output.
//
// The value output is backed by two channels, one for writing output, another
// for the back-chanel signal that the reader of the channel has gone.
type ValueOutput interface {
	// Outputs a value. Returns errs.ReaderGone if the reader is gone.
	Put(v interface{}) error
}

type valueOutput struct {
	data      chan<- interface{}
	sendStop  <-chan struct{}
	sendError *error
}

func (vo valueOutput) Put(v interface{}) error {
	select {
	case vo.data <- v:
		return nil
	case <-vo.sendStop:
		return *vo.sendError
	}
}

// ByteOutput defines the interface through which builtin commands access the
// byte output.
//
// It is a thin wrapper around the underlying *os.File value, only exposing
// the necessary methods for writing bytes and strings, and converting any
// syscall.EPIPE errors to errs.ReaderGone.
type ByteOutput interface {
	io.Writer
	io.StringWriter
}

type byteOutput struct {
	f *os.File
}

func (bo byteOutput) Write(p []byte) (int, error) {
	n, err := bo.f.Write(p)
	return n, convertReaderGone(err)
}

func (bo byteOutput) WriteString(s string) (int, error) {
	n, err := bo.f.WriteString(s)
	return n, convertReaderGone(err)
}

func convertReaderGone(err error) error {
	if pathErr, ok := err.(*os.PathError); ok {
		if pathErr.Err == epipe {
			return errs.ReaderGone{}
		}
	}
	return err
}

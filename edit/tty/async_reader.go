package tty

import (
	"bytes"
	"fmt"
	"os"
	"syscall"

	"github.com/elves/elvish/sys"
)

const (
	asyncReaderChanSize int = 128
)

// AsyncReader delivers a Unix fd stream to a channel of runes.
type AsyncReader struct {
	rd           *os.File
	rCtrl, wCtrl *os.File
	ctrlCh       chan struct{}
	ch           chan rune
	errCh        chan error
	debug        bool
}

// NewAsyncReader creates a new AsyncReader from a file.
func NewAsyncReader(rd *os.File) *AsyncReader {
	rCtrl, wCtrl, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return &AsyncReader{
		rd,
		rCtrl, wCtrl,
		make(chan struct{}),
		make(chan rune, asyncReaderChanSize),
		make(chan error),
		false,
	}
}

// Chan returns a channel onto which the AsyncReader writes the runes it reads.
func (ar *AsyncReader) Chan() <-chan rune {
	return ar.ch
}

// ErrorChan returns a channel onto which the AsyncReader writes the errors it
// encounters.
func (ar *AsyncReader) ErrorChan() <-chan error {
	return ar.errCh
}

// Run runs the AsyncReader. It blocks until Quit is called and should be
// called in a separate goroutine.
func (ar *AsyncReader) Run() {
	fd := ar.rd.Fd()
	cfd := ar.rCtrl.Fd()
	poller := sys.Poller{}
	var cBuf [1]byte

	if err := poller.Init([]uintptr{fd, cfd}, []uintptr{}); err != nil {
		ar.writeErrorAndWaitForQuit(err)
		return
	}

	for {
		rfds, _, err := poller.Poll(nil)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			ar.writeErrorAndWaitForQuit(err)
			return
		}
		for _, rfd := range *rfds {
			if rfd == cfd {
				// Consume the written byte
				ar.rCtrl.Read(cBuf[:])
				<-ar.ctrlCh
				return
			}
		}

		var buf [1]byte
		nr, err := syscall.Read(int(fd), buf[:])
		if nr != 1 {
			continue
		} else if err != nil {
			ar.writeErrorAndWaitForQuit(err)
			return
		}

		leader := buf[0]
		var (
			r       rune
			pending int
		)
		switch {
		case leader>>7 == 0:
			r = rune(leader)
		case leader>>5 == 0x6:
			r = rune(leader & 0x1f)
			pending = 1
		case leader>>4 == 0xe:
			r = rune(leader & 0xf)
			pending = 2
		case leader>>3 == 0x1e:
			r = rune(leader & 0x7)
			pending = 3
		}
		if ar.debug {
			fmt.Printf("leader 0x%x, pending %d, r = 0x%x\n", leader, pending, r)
		}
		for i := 0; i < pending; i++ {
			nr, err := syscall.Read(int(fd), buf[:])
			if nr != 1 {
				r = 0xfffd
				break
			} else if err != nil {
				ar.writeErrorAndWaitForQuit(err)
				return
			}
			r = r<<6 + rune(buf[0]&0x3f)
			if ar.debug {
				fmt.Printf("  got 0x%d, r = 0x%x\n", buf[0], r)
			}
		}

		// Write rune to ch, unless termination is requested.
		select {
		case ar.ch <- r:
		case <-ar.ctrlCh:
			ar.rCtrl.Read(cBuf[:])
			return
		}
	}
}

func subsequentBytes(b byte) int {
	i := 0
	for (b & 0x80) == 0x80 {
		i++
		b <<= 1
	}
	return i
}

func hexSeq(a []byte) string {
	var buf bytes.Buffer
	for i, b := range a {
		if i == 0 {
			buf.WriteRune('[')
		} else {
			buf.WriteRune(' ')
		}
		fmt.Fprintf(&buf, "0x%02x", b)
	}
	buf.WriteRune(']')
	return buf.String()
}

func (ar *AsyncReader) writeErrorAndWaitForQuit(err error) {
	var cBuf [1]byte

	select {
	case ar.errCh <- err:
	case <-ar.ctrlCh:
		ar.rCtrl.Read(cBuf[:])
		return
	}
	<-ar.ctrlCh
	ar.rCtrl.Read(cBuf[:])
}

// Quit terminates the loop of Run.
func (ar *AsyncReader) Quit() {
	_, err := ar.wCtrl.Write([]byte{'q'})
	if err != nil {
		panic(err)
	}
	ar.ctrlCh <- struct{}{}
}

// Close releases files and channels associated with the AsyncReader. It does
// not close the file used to create it.
func (ar *AsyncReader) Close() {
	ar.rCtrl.Close()
	ar.wCtrl.Close()
	close(ar.ctrlCh)
	close(ar.ch)
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

package tty

import (
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

	// unix-specific, no equavilent on windows
	if nonblock, _ := sys.GetNonblock(int(fd)); !nonblock {
		sys.SetNonblock(int(fd), true)
		defer sys.SetNonblock(int(fd), false)
	}

	if err := poller.Init([]uintptr{fd, cfd}, []uintptr{}); err != nil {
		// fatal error, unable to initialize poller
		ar.waitForQuit(err)
		return
	}

	for {
		rfds, _, err := poller.Poll(nil)
		if err != nil {
			switch err {
			case syscall.EINTR:
				continue
			default:
				ar.waitForQuit(err)
				return
			}
		}
		for _, rfd := range *rfds {
			if rfd == cfd {
				// Consume the written byte
				ar.rCtrl.Read(cBuf[:])
				<-ar.ctrlCh
				return
			}
		}

		bytes := make([]byte, 0, 32)
	ReadRunes:
		for {
			buf := make([]byte, 32)
			nr, err := syscall.Read(int(fd), buf[:])

			if err == nil {
				bytes = append(bytes, buf[:nr]...)
			} else {
				if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
					// All input read, break the loop.
					break ReadRunes
				}
				// Write error to errCh, unless termination is requested.
				select {
				case ar.errCh <- err:
				case <-ar.ctrlCh:
					ar.rCtrl.Read(cBuf[:])
					return
				}
			}
		}
		// TODO(xiaq): Invalid UTF-8 will result in a bunch of \ufffd, which is
		// not helpful for debugging.
		for _, r := range string(bytes) {
			// Write error to ch, unless termination is requested.
			select {
			case ar.ch <- r:
			case <-ar.ctrlCh:
				ar.rCtrl.Read(cBuf[:])
				return
			}
		}
	}
}

func (ar *AsyncReader) waitForQuit(err error) {
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

package tty

import (
	"bufio"
	"io"
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
	bufrd        *bufio.Reader
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
		bufio.NewReaderSize(rd, 0),
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
		// TODO show erorr
		return
	}

	for {
		rfds, _, err := poller.Poll(nil)
		if err != nil {
			switch err {
			case syscall.EINTR:
				continue
			default:
				ar.errCh <- err
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
	ReadRune:
		for {
			r, _, err := ar.bufrd.ReadRune()
			switch err {
			case nil:
				// Logger.Printf("read rune: %q", r)
				select {
				case ar.ch <- r:
				case <-ar.ctrlCh:
					ar.rCtrl.Read(cBuf[:])
					return
				}
			case io.EOF:
				return
			default:
				// BUG(xiaq): AsyncReader relies on the undocumented fact
				// that (*os.File).Read returns an *os.File.PathError
				patherr, ok := err.(*os.PathError) //.Err
				if ok && patherr.Err == syscall.EWOULDBLOCK || patherr.Err == syscall.EAGAIN {
					break ReadRune
				} else {
					select {
					case ar.errCh <- err:
					case <-ar.ctrlCh:
						ar.rCtrl.Read(cBuf[:])
						return
					}
				}
			}
		}
	}
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

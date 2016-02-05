package edit

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
}

func NewAsyncReader(rd *os.File) *AsyncReader {
	ar := &AsyncReader{
		rd:     rd,
		bufrd:  bufio.NewReaderSize(rd, 0),
		ctrlCh: make(chan struct{}),
		ch:     make(chan rune, asyncReaderChanSize),
	}

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	ar.rCtrl, ar.wCtrl = r, w
	return ar
}

func (ar *AsyncReader) Chan() <-chan rune {
	return ar.ch
}

func (ar *AsyncReader) Run() {
	fd := int(ar.rd.Fd())
	cfd := int(ar.rCtrl.Fd())
	maxfd := max(fd, cfd)
	fs := sys.NewFdSet()
	var cBuf [1]byte

	if nonblock, _ := sys.GetNonblock(fd); !nonblock {
		sys.SetNonblock(fd, true)
		defer sys.SetNonblock(fd, false)
	}

	for {
		fs.Set(fd, cfd)
		err := sys.Select(maxfd+1, fs, nil, nil, nil)
		if err != nil {
			switch err {
			case syscall.EINTR:
				continue
			default:
				panic(err)
			}
		}
		if fs.IsSet(cfd) {
			// Consume the written byte
			ar.rCtrl.Read(cBuf[:])
			<-ar.ctrlCh
			return
		} else {
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
					if !ok {
						panic(err)
					}
					e := patherr.Err
					if e == syscall.EWOULDBLOCK || e == syscall.EAGAIN {
						break ReadRune
					} else {
						panic(err)
					}
				}
			}
		}
	}
}

func (ar *AsyncReader) Quit() {
	_, err := ar.wCtrl.Write([]byte{'q'})
	if err != nil {
		panic(err)
	}
	ar.ctrlCh <- struct{}{}
}

func (ar *AsyncReader) Close() {
	ar.rCtrl.Close()
	ar.wCtrl.Close()
	close(ar.ctrlCh)
	close(ar.ch)
}

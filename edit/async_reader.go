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
	ackCtrl      chan bool // Used to synchronize receiving of ctrl message
	ch           chan rune
}

func NewAsyncReader(rd *os.File) *AsyncReader {
	ar := &AsyncReader{
		rd:      rd,
		bufrd:   bufio.NewReaderSize(rd, 0),
		ackCtrl: make(chan bool),
		ch:      make(chan rune, asyncReaderChanSize),
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

func (ar *AsyncReader) Start() {
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
			ar.ackCtrl <- true
			return
		} else {
			r, _, err := ar.bufrd.ReadRune()
			switch err {
			case nil:
				ar.ch <- r
			case io.EOF:
				return
			default:
				// BUG(xiaq): AsyncReader relies on the undocumented fact
				// that (*os.File).Read returns an *os.File.PathError
				e := err.(*os.PathError).Err
				if e != syscall.EWOULDBLOCK && e != syscall.EAGAIN {
					panic(err)
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
	select {
	case <-ar.ch:
	default:
	}
	<-ar.ackCtrl
}

func (ar *AsyncReader) Close() {
	ar.rCtrl.Close()
	ar.wCtrl.Close()
	close(ar.ackCtrl)
	close(ar.ch)
}

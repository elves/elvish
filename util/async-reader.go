package util

import (
	"bufio"
	"io"
	"os"
	"syscall"

	"github.com/xiaq/elvish/sys"
)

const (
	asyncReaderChanSize int = 128
)

const (
	asyncReaderStop     byte = 's'
	asyncReaderContinue      = 'c'
	asyncReaderQuit          = 'q'
)

// AsyncReader delivers a Unix fd stream to a channel of runes.
type AsyncReader struct {
	rd           *os.File
	bufrd        *bufio.Reader
	rCtrl, wCtrl *os.File
	ch           chan rune
}

func NewAsyncReader(rd *os.File) *AsyncReader {
	ar := &AsyncReader{
		rd:    rd,
		bufrd: bufio.NewReaderSize(rd, 0),
		ch:    make(chan rune, asyncReaderChanSize),
	}

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	ar.rCtrl, ar.wCtrl = r, w
	go ar.run()
	return ar
}

func (ar *AsyncReader) Chan() <-chan rune {
	return ar.ch
}

func (ar *AsyncReader) run() {
	fd := int(ar.rd.Fd())
	cfd := int(ar.rCtrl.Fd())
	maxfd := MaxInt(fd, cfd)
	fs := sys.NewFdSet()
	var cBuf [1]byte

	defer close(ar.ch)

	for {
		fs.Set(fd, cfd)
		_, err := sys.Select(maxfd+1, fs, nil, nil, nil)
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
			switch cBuf[0] {
			case asyncReaderQuit:
				return
			case asyncReaderStop:
			Stop:
				for {
					ar.rCtrl.Read(cBuf[:])
					switch cBuf[0] {
					case asyncReaderQuit:
						return
					case asyncReaderContinue:
						break Stop
					}
				}
			}
		} else {
		ReadRune:
			for {
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

func (ar *AsyncReader) ctrl(r byte) {
	_, err := ar.wCtrl.Write([]byte{r})
	if err != nil {
		panic(err)
	}
}

func (ar *AsyncReader) Stop() {
	ar.ctrl(asyncReaderStop)
}

func (ar *AsyncReader) Continue() {
	ar.ctrl(asyncReaderContinue)
}

func (ar *AsyncReader) Quit() {
	ar.ctrl(asyncReaderQuit)
}

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

// AsyncReader delivers a Unix fd stream to a channel of runes.
type AsyncReader struct {
	rd           *os.File
	bufrd        *bufio.Reader
	rQuit, wQuit *os.File
}

func NewAsyncReader(rd *os.File) *AsyncReader {
	ar := &AsyncReader{
		rd:    rd,
		bufrd: bufio.NewReaderSize(rd, 0),
	}

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	ar.rQuit, ar.wQuit = r, w
	return ar
}

func (ar *AsyncReader) run(ch chan<- rune) {
	fd := int(ar.rd.Fd())
	qfd := int(ar.rQuit.Fd())
	maxfd := MaxInt(fd, qfd)
	fs := sys.NewFdSet()
	var qBuf [1]byte

	defer close(ch)

	for {
		fs.Set(fd, qfd)
		_, err := sys.Select(maxfd+1, fs, nil, nil, nil)
		if err != nil {
			switch err {
			case syscall.EINTR:
				continue
			default:
				panic(err)
			}
		}
		if fs.IsSet(qfd) {
			// Consume the written byte
			ar.rQuit.Read(qBuf[:])
			return
		} else {
		ReadRune:
			for {
				r, _, err := ar.bufrd.ReadRune()
				switch err {
				case nil:
					ch <- r
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

func (ar *AsyncReader) Start() <-chan rune {
	ch := make(chan rune, asyncReaderChanSize)
	go ar.run(ch)
	return ch
}

func (ar *AsyncReader) Stop() {
	_, err := ar.wQuit.Write([]byte("x"))
	if err != nil {
		panic(err)
	}
}

package async

import (
	"io"
	"time"
	"bufio"
)

// ReadRuneRet packs the return value of (*bufio.Reader).ReadRune.
type ReadRuneRet struct {
	r rune
	size int
	err error
}

// RuneReader wraps bufio.Reader to support ReadRune with timeout.
type RuneReader struct {
	*bufio.Reader
	rets chan ReadRuneRet
}

func NewRuneReader(r io.Reader) *RuneReader {
	rr := RuneReader{bufio.NewReaderSize(r, 0), make(chan ReadRuneRet)}
	go rr.serve()
	return &rr
}

func (rr *RuneReader) serve() {
	for {
		r, size, err := rr.Reader.ReadRune()
		rr.rets <- ReadRuneRet{r, size, err}
	}
}

// ReadRuneTimeout is like ReadRune but blocks for at most d. If there was no
// rune read, err is set to Timeout.
func (rr *RuneReader) ReadRuneTimeout(d time.Duration) (r rune, size int, err error) {
	select {
	case rt := <-rr.rets:
		return rt.r, rt.size, rt.err
	case <- after(d):
		return 0, 0, Timeout
	}
}

func (rr *RuneReader) ReadRune() (rune, int, error) {
	return rr.ReadRuneTimeout(0)
}

// after is like time.After but d == 0 returns a channel that is never sent
// to.
func after(d time.Duration) <-chan time.Time {
	if d > 0 {
		return time.After(d)
	}
	return make(chan time.Time)
}

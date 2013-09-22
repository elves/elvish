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
	Timeout time.Duration
	rets chan ReadRuneRet
}

func NewRuneReader(r io.Reader, d time.Duration) *RuneReader {
	rr := RuneReader{bufio.NewReaderSize(r, 0), d, make(chan ReadRuneRet)}
	go rr.serve()
	return &rr
}

func (rr *RuneReader) serve() {
	for {
		r, size, err := rr.Reader.ReadRune()
		rr.rets <- ReadRuneRet{r, size, err}
	}
}

func (rr *RuneReader) ReadRune() (r rune, size int, err error) {
	select {
	case rt := <-rr.rets:
		return rt.r, rt.size, rt.err
	case <- after(rr.Timeout):
		return 0, 0, Timeout
	}
}

// after is like time.After but d == 0 returns a channel that is never sent
// to.
func after(d time.Duration) <-chan time.Time {
	if d > 0 {
		return time.After(d)
	}
	return make(chan time.Time)
}

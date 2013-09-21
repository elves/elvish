package async

import (
	"io"
	"fmt"
	"bufio"
)

type Item struct {
	rune
	Err error
}

func (it Item) GoString() string {
	return fmt.Sprintf("async.Item{rune: %q, Err: %v}", it.rune, it.Err)
}

type RuneReader struct {
	reader *bufio.Reader
	Items chan Item
	Go chan bool
}

func NewRuneReaderSize(r *bufio.Reader, n int) *RuneReader {
	rr := RuneReader{r, make(chan Item, n), make(chan bool)}
	go rr.serve()
	return &rr
}

func NewRuneReader(r *bufio.Reader) *RuneReader {
	return NewRuneReaderSize(r, 0)
}

func (rr *RuneReader) serve() {
	for {
		if !<-rr.Go {
			break
		}
		r, _, err := rr.reader.ReadRune()
		if err == io.EOF {
			break
		}
		rr.Items <- Item{r, err}
	}
	close(rr.Items)
	close(rr.Go)
	return
}

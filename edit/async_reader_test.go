package edit

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// Pretty arbitrary numbers. May not reveal deadlocks on all machines.

var (
	NWrite  = 40960
	Run     = 6400
	Timeout = 500 * time.Millisecond
)

func f(done chan struct{}) {
	r, w, _ := os.Pipe()
	defer r.Close()
	defer w.Close()
	ar := NewAsyncReader(r)
	defer ar.Close()
	fmt.Fprintf(w, "%*s", NWrite, "")
	go ar.Start()
	ar.Quit()
	done <- struct{}{}
}

func TestAsyncReaderDeadlock(t *testing.T) {
	done := make(chan struct{})
	for i := 0; i < Run; i++ {
		if i%100 == 0 {
			fmt.Printf("\r%d/%d", i, Run)
		}
		go f(done)
		select {
		case <-done:
			// no deadlock trigerred
		case <-time.After(Timeout):
			// deadlock
			t.Fatalf("AsyncReader deadlock trigerred on run %d/%d", i, Run)
		}
	}
}

package edit

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/elves/elvish/sys"
)

// Pretty arbitrary numbers. May not reveal deadlocks on all machines.

var (
	NWrite    = 1024
	Run       = 1024
	Timeout   = 500 * time.Millisecond
	MaxJitter = time.Millisecond
)

func jitter() {
	time.Sleep(time.Duration(float64(MaxJitter) * rand.Float64()))
}

func f(done chan struct{}) {
	r, w, _ := os.Pipe()
	defer r.Close()
	defer w.Close()
	ar := NewAsyncReader(r)
	defer ar.Close()
	fmt.Fprintf(w, "%*s", NWrite, "")
	go func() {
		jitter()
		ar.Start()
	}()
	jitter()
	ar.Quit()
	done <- struct{}{}
}

func TestAsyncReaderDeadlock(t *testing.T) {
	done := make(chan struct{})
	isatty := sys.IsATTY(1)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < Run; i++ {
		if isatty {
			fmt.Printf("\r%d/%d", i, Run)
		}
		go f(done)
		select {
		case <-done:
			// no deadlock trigerred
		case <-time.After(Timeout):
			// deadlock
			t.Errorf("%s", sys.DumpStack())
			t.Fatalf("AsyncReader deadlock trigerred on run %d/%d, stack trace:\n%s", i, Run, sys.DumpStack())
		}
	}
}

package tty

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
	DeadlockNWrite    = 1024
	DeadlockRun       = 64
	DeadlockTimeout   = 500 * time.Millisecond
	DeadlockMaxJitter = time.Millisecond
)

func jitter() {
	time.Sleep(time.Duration(float64(DeadlockMaxJitter) * rand.Float64()))
}

func f(done chan struct{}) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer r.Close()
	defer w.Close()

	ar := NewAsyncReader(r)
	defer ar.Close()
	fmt.Fprintf(w, "%*s", DeadlockNWrite, "")
	go func() {
		jitter()
		ar.Run()
	}()
	jitter()
	ar.Quit()
	done <- struct{}{}
}

func TestAsyncReaderDeadlock(t *testing.T) {
	done := make(chan struct{})
	isatty := sys.IsATTY(1)
	rand.Seed(time.Now().UTC().UnixNano())

	timer := time.NewTimer(DeadlockTimeout)
	for i := 0; i < DeadlockRun; i++ {
		if isatty {
			fmt.Printf("\r%d/%d ", i+1, DeadlockRun)
		}

		go f(done)

		select {
		case <-done:
			// no deadlock trigerred
		case <-timer.C:
			// deadlock
			t.Errorf("%s", sys.DumpStack())
			t.Fatalf("AsyncReader deadlock trigerred on run %d/%d, stack trace:\n%s", i, DeadlockRun, sys.DumpStack())
		}
		timer.Reset(DeadlockTimeout)
	}
	if isatty {
		fmt.Print("\r       \r")
	}
}

var ReadTimeout = time.Second

func TestAsyncReader(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer r.Close()
	defer w.Close()

	ar := NewAsyncReader(r)
	defer ar.Close()
	go ar.Run()

	go func() {
		var i rune
		for i = 0; i <= 1280; i += 10 {
			w.WriteString(string(i))
		}
	}()

	var i rune
	timer := time.NewTimer(ReadTimeout)
	for i = 0; i <= 1280; i += 10 {
		select {
		case r := <-ar.Chan():
			if r != i {
				t.Fatalf("expect %q, got %q\n", i, r)
			}
		case <-timer.C:
			t.Fatalf("read timeout (i = %d)", i)
		}
		timer.Reset(ReadTimeout)
	}
	ar.Quit()
}

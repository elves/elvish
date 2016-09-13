package stub

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func testSignal(t *testing.T, stub *Stub, sig syscall.Signal) {
	stub.Process().Signal(sig)
	select {
	case gotsig := <-stub.Signals():
		if gotsig != sig {
			t.Errorf("got %v, want %v", gotsig, sig)
		}
	case <-time.After(time.Millisecond * 10):
		t.Errorf("signal not relayed after 10ms")
	}
}

func TestStub(t *testing.T) {
	stub, err := NewStub(os.Stderr)
	if err != nil {
		t.Skip(err)
	}

	// Non-INT signals should be relayed onto Signals, but not IntSignals.
	testSignal(t, stub, syscall.SIGUSR1)
	select {
	case <-stub.IntSignals():
		t.Errorf("SIGUSR1 relayed onto IntSignals")
	case <-time.After(time.Millisecond):
	}

	// INT signals should be relayed onto both Signals and IntSignals.
	testSignal(t, stub, syscall.SIGINT)
	select {
	case <-stub.IntSignals():
	case <-time.After(10 * time.Millisecond):
		t.Errorf("SIGINT not relayed onto IntSignals within 10ms")
	}

	// Setting title and dir of the stub shouldn't cause the stub to terminate,
	// even if the payload is invalid or contains newlines.
	stub.SetTitle("x\ny")
	stub.Chdir("/xyz/haha")
	select {
	case <-stub.State():
		t.Errorf("stub exited prematurely")
	default:
	}

	// Calling Terminate should really terminate the process.
	stub.Terminate()
	select {
	case <-stub.State():
	case <-time.After(time.Millisecond * 10):
		t.Errorf("stub didn't exit within 10ms")
	}
}

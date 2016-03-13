package stub

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestStub(t *testing.T) {
	stub, err := NewStub(os.Stderr)
	if err != nil {
		t.Skip(err)
	}
	proc := stub.Process()

	// Signals should be relayed.
	proc.Signal(syscall.SIGINT)
	select {
	case sig := <-stub.Signals():
		if sig != syscall.SIGINT {
			t.Errorf("got %v, want SIGINT", sig)
		}
	case <-time.After(time.Millisecond * 10):
		t.Errorf("signal not relayed after 10ms")
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

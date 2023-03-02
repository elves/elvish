//go:build unix

package daemon

import (
	"os"
	"syscall"
	"testing"
)

func TestProgram_QuitsOnSystemSignal_SIGINT(t *testing.T) {
	testProgram_QuitsOnSystemSignal(t, syscall.SIGINT)
}

func TestProgram_QuitsOnSystemSignal_SIGTERM(t *testing.T) {
	testProgram_QuitsOnSystemSignal(t, syscall.SIGTERM)
}

func testProgram_QuitsOnSystemSignal(t *testing.T, sig os.Signal) {
	t.Helper()
	setup(t)
	startServerOpts(t, cli("sock", "db"), ServeOpts{Signals: nil})
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}
	p.Signal(sig)
	// startServerOpts will wait for server to terminate at cleanup
}

//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

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
	startServerSigCh(t, cli("sock", "db"), nil)
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}
	p.Signal(sig)
	// startServerSigCh will wait for server to terminate at cleanup
}

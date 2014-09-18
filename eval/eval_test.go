package eval

import (
	"strconv"
	"syscall"
	"testing"
)

func TestNewEvaluator(t *testing.T) {
	ev := NewEvaluator()
	pid := strconv.Itoa(syscall.Getpid())
	if (*ev.scope["pid"]).String() != pid {
		t.Errorf(`ev.scope["pid"] = %v, want %v`, ev.scope["pid"], pid)
	}
}

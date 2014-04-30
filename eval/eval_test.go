package eval

import (
	"strconv"
	"syscall"
	"testing"
)

func strsEqual(s1 []string, s2 []string) bool {
	if len(s1) == len(s2) {
		for i := range s1 {
			if s1[i] != s2[i] {
				return false
			}
			return true
		}
	}
	return false
}

func TestNewEvaluator(t *testing.T) {
	ev := NewEvaluator()
	pid := strconv.Itoa(syscall.Getpid())
	if (*ev.scope["pid"]).String() != pid {
		t.Errorf(`ev.scope["pid"] = %v, want %v`, ev.scope["pid"], pid)
	}
}

// +build !windows

package sys

import (
	"syscall"
	"testing"
)

func TestPoller(t *testing.T) {
	var p1, p2 [2]int
	mustNil(syscall.Pipe(p1[:]))
	mustNil(syscall.Pipe(p2[:]))
	poller := Poller{}
	poller.Init([]uintptr{uintptr(p1[0]), uintptr(p2[0])}, []uintptr{})
	go func() {
		syscall.Write(p1[1], []byte("to p1"))
		syscall.Write(p2[1], []byte("to p2"))
		syscall.Close(p1[1])
		syscall.Close(p2[1])
	}()
	_, _, e := poller.Poll(nil)
	if e != nil {
		t.Errorf("Poll(nil) => %v, want <nil>", e)
	}
	syscall.Close(p1[0])
	syscall.Close(p2[0])
}

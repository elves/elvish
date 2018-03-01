// +build !windows

package sys

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestFdSet(t *testing.T) {
	fs := NewFdSet(42, 233)
	fs.Set(77)
	fds := []int{42, 233, 77}
	for _, i := range fds {
		if !fs.IsSet(i) {
			t.Errorf("fs.IsSet(%d) => false, want true", i)
		}
	}
	fs.Clear(233)
	if fs.IsSet(233) {
		t.Errorf("fs.IsSet(233) => true, want false")
	}
	fs.Zero()
	for _, i := range fds {
		if fs.IsSet(i) {
			t.Errorf("fs.IsSet(%d) => true, want false", i)
		}
	}
}

func TestSelect(t *testing.T) {
	var p1, p2 [2]int
	mustNil(unix.Pipe(p1[:]))
	mustNil(unix.Pipe(p2[:]))
	fs := NewFdSet(p1[0], p2[0])
	var maxfd int
	if p1[0] > p2[0] {
		maxfd = p1[0] + 1
	} else {
		maxfd = p2[0] + 1
	}
	go func() {
		unix.Write(p1[1], []byte("to p1"))
		unix.Write(p2[1], []byte("to p2"))
		unix.Close(p1[1])
		unix.Close(p2[1])
	}()
	e := Select(maxfd+1, fs, nil, nil)
	if e != nil {
		t.Errorf("Select(%v, %v, nil, nil) => %v, want <nil>",
			maxfd+1, fs, e)
	}
	unix.Close(p1[0])
	unix.Close(p2[0])
}

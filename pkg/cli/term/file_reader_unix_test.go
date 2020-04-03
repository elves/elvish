// +build !windows,!plan9

package term

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/elves/elvish/pkg/testutil"
)

func TestFileReader_ReadByteWithTimeout(t *testing.T) {
	r, w, cleanup := setupFileReader()
	defer cleanup()

	content := []byte("0123456789")
	w.Write(content)

	// Test successful ReadByteWithTimeout calls.
	for i := 0; i < len(content); i++ {
		t.Run(fmt.Sprintf("byte %d", i), func(t *testing.T) {
			b, err := r.ReadByteWithTimeout(-1)
			if err != nil {
				t.Errorf("got err %v, want nil", err)
			}
			if b != content[i] {
				t.Errorf("got byte %q, want %q", b, content[i])
			}
		})
	}
}

func TestFileReader_ReadByteWithTimeout_EOF(t *testing.T) {
	r, w, cleanup := setupFileReader()
	defer cleanup()

	w.Close()
	_, err := r.ReadByteWithTimeout(-1)
	if err != io.EOF {
		t.Errorf("got byte %v, want %v", err, io.EOF)
	}
}

func TestFileReader_ReadByteWithTimeout_Timeout(t *testing.T) {
	r, _, cleanup := setupFileReader()
	defer cleanup()

	_, err := r.ReadByteWithTimeout(testutil.ScaledMs(1))
	if err != errTimeout {
		t.Errorf("got err %v, want %v", err, errTimeout)
	}
}

func TestFileReader_Stop(t *testing.T) {
	r, _, cleanup := setupFileReader()
	defer cleanup()

	errCh := make(chan error, 1)
	go func() {
		_, err := r.ReadByteWithTimeout(-1)
		errCh <- err
	}()
	r.Stop()

	if err := <-errCh; err != ErrStopped {
		t.Errorf("got err %v, want %v", err, ErrStopped)
	}
}

func setupFileReader() (reader fileReader, writer *os.File, cleanup func()) {
	pr, pw, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	r, err := newFileReader(pr)
	if err != nil {
		panic(err)
	}
	return r, pw, func() {
		r.Close()
		pr.Close()
		pw.Close()
	}
}

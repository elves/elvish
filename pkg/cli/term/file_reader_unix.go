// +build !windows,!plan9

package term

import (
	"errors"
	"io"
	"os"
	"syscall"

	"github.com/elves/elvish/pkg/sys"
)

var errStopped = errors.New("stopped")

// A helper for reading from a file.
type fileReader interface {
	// ReadByte reads a single byte from the underlying file. May return
	// errStopped to signal that Stop was called during the read.
	ReadByte() (byte, error)
	// Stop stops any outstanding read.
	Stop() error
	// Close releases new resources allocated for the fileReader. It does not
	// close the underlying file.
	Close()
}

func newFileReader(file *os.File) (fileReader, error) {
	rStop, wStop, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	return &bReader{file, rStop, wStop}, nil
}

type bReader struct {
	file  *os.File
	rStop *os.File
	wStop *os.File
}

const maxNoProgress = 10

func (r *bReader) ReadByte() (byte, error) {
	for {
		ready, err := sys.WaitForRead(r.file, r.rStop)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return 0, err
		}
		if ready[1] {
			return 0, errStopped
		}
		var b [1]byte
		nr, err := r.file.Read(b[:])
		if err != nil {
			return 0, err
		}
		if nr != 1 {
			return 0, io.ErrNoProgress
		}
		return b[0], nil
	}
}

func (r *bReader) Stop() error {
	_, err := r.wStop.Write([]byte{'q'})
	return err
}

func (r *bReader) Close() {
	r.rStop.Close()
	r.wStop.Close()
}

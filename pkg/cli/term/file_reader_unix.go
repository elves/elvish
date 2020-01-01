// +build !windows,!plan9

package term

import (
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/elves/elvish/pkg/sys"
)

// A helper for reading from a file.
type fileReader interface {
	byteReaderWithTimeout
	// Stop stops any outstanding read call. It blocks until the read returns.
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
	return &bReader{file: file, rStop: rStop, wStop: wStop}, nil
}

type bReader struct {
	file  *os.File
	rStop *os.File
	wStop *os.File
	// A mutex that is held when Read is in process.
	mutex sync.Mutex
}

const maxNoProgress = 10

func (r *bReader) ReadByteWithTimeout(timeout time.Duration) (byte, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for {
		ready, err := sys.WaitForRead(timeout, r.file, r.rStop)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return 0, err
		}
		if ready[1] {
			var b [1]byte
			r.rStop.Read(b[:])
			return 0, ErrStopped
		}
		if !ready[0] {
			return 0, errTimeout
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
	r.mutex.Lock()
	r.mutex.Unlock()
	return err
}

func (r *bReader) Close() {
	r.rStop.Close()
	r.wStop.Close()
}

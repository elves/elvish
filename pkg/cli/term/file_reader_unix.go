//go:build unix

package term

import (
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"src.elv.sh/pkg/sys/eunix"
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

func (r *bReader) ReadByteWithTimeout(timeout time.Duration) (byte, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for {
		ready, err := eunix.WaitForRead(timeout, r.file, r.rStop)
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
	//lint:ignore SA2001 We only lock the mutex to make sure that
	// ReadByteWithTimeout has exited, so we unlock it immediately.
	r.mutex.Unlock()
	return err
}

func (r *bReader) Close() {
	r.rStop.Close()
	r.wStop.Close()
}

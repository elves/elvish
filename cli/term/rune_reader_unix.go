// +build !windows,!plan9

package tty

import (
	"fmt"
	"os"
	"syscall"

	"github.com/elves/elvish/sys"
)

const (
	runeReaderChanSize int = 128
)

// runeReader reads a Unix file continuously, assemble the bytes it reads into
// runes (assuming UTF-8), and delivers them on a channel.
type runeReader struct {
	file      *os.File
	rStop     *os.File
	wStop     *os.File
	stopChan  chan struct{}
	runeChan  chan rune
	errorChan chan error
	debug     bool
}

// newRuneReader creates a new runeReader from a file.
func newRuneReader(file *os.File) *runeReader {
	rStop, wStop, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return &runeReader{
		file,
		rStop, wStop,
		make(chan struct{}),
		make(chan rune, runeReaderChanSize),
		make(chan error),
		false,
	}
}

// Chan returns a channel onto which the runeReader writes the runes it reads.
func (ar *runeReader) Chan() <-chan rune {
	return ar.runeChan
}

// ErrorChan returns a channel onto which the runeReader writes the errors it
// encounters.
func (ar *runeReader) ErrorChan() <-chan error {
	return ar.errorChan
}

// Start starts the runeReader.
func (ar *runeReader) Start() {
	go ar.run()
}

// run runs the runeReader. It blocks until Quit is called and should be called
// in a separate goroutine.
func (ar *runeReader) run() {
	var buf [1]byte

	for {
		ready, err := sys.WaitForRead(ar.file, ar.rStop)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			ar.fatal(err)
			return
		}
		if ready[1] {
			// Consume the written byte
			ar.rStop.Read(buf[:])
			<-ar.stopChan
			return
		}

		nr, err := ar.file.Read(buf[:])
		if nr != 1 {
			continue
		} else if err != nil {
			ar.fatal(err)
			return
		}

		leader := buf[0]
		var (
			r       rune
			pending int
		)
		switch {
		case leader>>7 == 0:
			r = rune(leader)
		case leader>>5 == 0x6:
			r = rune(leader & 0x1f)
			pending = 1
		case leader>>4 == 0xe:
			r = rune(leader & 0xf)
			pending = 2
		case leader>>3 == 0x1e:
			r = rune(leader & 0x7)
			pending = 3
		}
		if ar.debug {
			fmt.Printf("leader 0x%x, pending %d, r = 0x%x\n", leader, pending, r)
		}
		for i := 0; i < pending; i++ {
			nr, err := ar.file.Read(buf[:])
			if nr != 1 {
				r = 0xfffd
				break
			} else if err != nil {
				ar.fatal(err)
				return
			}
			r = r<<6 + rune(buf[0]&0x3f)
			if ar.debug {
				fmt.Printf("  got 0x%d, r = 0x%x\n", buf[0], r)
			}
		}

		// Write rune to ch, unless termination is requested.
		select {
		case ar.runeChan <- r:
		case <-ar.stopChan:
			ar.rStop.Read(buf[:])
			return
		}
	}
}

func (ar *runeReader) fatal(err error) {
	var cBuf [1]byte

	select {
	case ar.errorChan <- err:
	case <-ar.stopChan:
		ar.rStop.Read(cBuf[:])
		return
	}
	<-ar.stopChan
	ar.rStop.Read(cBuf[:])
}

// Stop terminates the loop of Run.
func (ar *runeReader) Stop() {
	_, err := ar.wStop.Write([]byte{'q'})
	if err != nil {
		panic(err)
	}
	ar.stopChan <- struct{}{}
}

// Close releases files and channels associated with the AsyncReader. It does
// not close the file used to create it.
func (ar *runeReader) Close() {
	ar.rStop.Close()
	ar.wStop.Close()
	close(ar.stopChan)
	close(ar.runeChan)
}

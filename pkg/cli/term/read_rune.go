//go:build unix

package term

import (
	"time"
)

type byteReaderWithTimeout interface {
	// ReadByteWithTimeout reads a single byte with a timeout. A negative
	// timeout means no timeout.
	ReadByteWithTimeout(timeout time.Duration) (byte, error)
}

const badRune = '\ufffd'

var utf8SeqTimeout = 10 * time.Millisecond

// Reads a rune from the reader. The timeout applies to the first byte; a
// negative value means no timeout.
func readRune(rd byteReaderWithTimeout, timeout time.Duration) (rune, error) {
	leader, err := rd.ReadByteWithTimeout(timeout)
	if err != nil {
		return badRune, err
	}
	var r rune
	pending := 0
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
	for i := 0; i < pending; i++ {
		b, err := rd.ReadByteWithTimeout(utf8SeqTimeout)
		if err != nil {
			return badRune, err
		}
		r = r<<6 + rune(b&0x3f)
	}
	return r, nil
}

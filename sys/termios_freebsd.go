// Copyright 2015 go-termios Author. All Rights Reserved.
// https://github.com/go-termios/termios
// Author: John Lenton <chipaca@github.com>

package sys

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// Apparently missing from sys/unix
const (
	// taken from sys/file.h
	FREAD  = 1
	FWRITE = 2

	// taken from termios.h
	TCOOFF = 1
	TCOON  = 2
	TCIOFF = 3
	TCION  = 4
)

const (
	getAttrIOCTL      = unix.TIOCGETA
	setAttrNowIOCTL   = unix.TIOCSETA
	setAttrDrainIOCTL = unix.TIOCSETAW
	setAttrFlushIOCTL = unix.TIOCSETAF
	flushIOCTL        = unix.TIOCFLUSH
)

const (
	// for MakeRaw()
	rawImaskOff = ^uint32(unix.IMAXBEL | unix.IXOFF | unix.INPCK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON | unix.IGNPAR)
	rawImaskOn  = uint32(unix.IGNBRK)
	rawOmaskOff = ^uint32(unix.OPOST)
	rawOmaskOn  = uint32(1<<32 - 1)
	rawLmaskOff = ^uint32(unix.ECHO | unix.ECHOE | unix.ECHOK | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN | unix.NOFLSH | unix.TOSTOP | unix.PENDIN)
	rawLmaskOn  = uint32(1<<32 - 1)
	rawCmaskOff = ^uint32(unix.CSIZE | unix.PARENB)
	rawCmaskOn  = uint32(unix.CS8 | unix.CREAD)
)

// GetSpeed gets the stored baud rate.
func (tio *Termios) GetSpeed() (in int, out int) {
	return int(tio.Ispeed), int(tio.Ospeed)
}

// SetSpeed sets the baud rate.
func (tio *Termios) SetSpeed(in int, out int) error {
	tio.Ispeed = uint32(in)
	tio.Ospeed = uint32(out)

	return nil
}

// SendBreak transmits a continuous stream of zeros for 0.4 seconds if
// duration is 0, or for the given number of deciseconds if not, if
// the terminal supports breaks.
func SendBreak(fd uintptr, duration int) error {
	if duration == 0 {
		duration = 4
	}
	if err := ioctlu(fd, unix.TIOCSBRK, 0); err != nil {
		return err
	}
	time.Sleep(time.Duration(duration) * time.Second / 10)
	return ioctlu(fd, unix.TIOCCBRK, 0)
}

// Drain waits until all output written to the terminal referenced by fd has
// been transmitted to the terminal.
func Drain(fd uintptr) error {
	return ioctlu(fd, unix.TIOCDRAIN, 0)
}

// Flow manages the suspending of data transmission or reception on the
// terminal referenced by fd.  The value of action must be one of the
// following:
// TCOOFF  Suspend output.
// TCOON   Restart suspended output.
// TCIOFF  Transmit a STOP character (the XOFF in XON/XOFF)
// TCION   Transmit a START character (the XON).
//
// The last two are not (yet) implemented; TODO: find out if doing
// them on FreeBSD really needs something like terminfo, as the
// “Special Control Characters” note in termios(4) seems to imply.
func Flow(fd uintptr, action int) error {
	var idx int

	switch action {
	case TCOOFF:
		return ioctlu(fd, unix.TIOCSTOP, 0)
	case TCOON:
		return ioctlu(fd, unix.TIOCSTART, 0)
	case TCIOFF:
		idx = unix.VSTOP
	case TCION:
		idx = unix.VSTART
	default:
		return ErrInvalidAction
	}

	tio, err := GetAttr(fd)
	if err != nil {
		return err
	}
	_, err = os.NewFile(fd, "").Write([]byte{tio.Cc[idx]})
	return err
}

func (q Queue) bits() uintptr {
	switch q {
	case InputQueue:
		return FREAD
	case OutputQueue:
		return FWRITE
	default:
		return 0
	}
}

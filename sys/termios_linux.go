// +build linux

// Copyright 2015 go-termios Author. All Rights Reserved.
// https://github.com/go-termios/termios
// Author: John Lenton <chipaca@github.com>

package sys

import (
	"fmt"
	"sort"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Linux-specific bit meanings, from bits/termios.h
const (
	// linux-specific c_cflag bits:
	CBAUD   = uint32(unix.CBAUD)
	CBAUDEX = uint32(unix.CBAUDEX)

// 	B500000  = uint32(unix.B500000)
// 	B576000  = uint32(unix.B576000)
// 	B1000000 = uint32(unix.B1000000)
// 	B1152000 = uint32(unix.B1152000)
// 	B1500000 = uint32(unix.B1500000)
// 	B2000000 = uint32(unix.B2000000)
// 	B2500000 = uint32(unix.B2500000)
// 	B3000000 = uint32(unix.B3000000)
// 	B3500000 = uint32(unix.B3500000)
// 	B4000000 = uint32(unix.B4000000)
// 	CIBAUD   = uint32(unix.CIBAUD)
// 	CMSPAR   = uint32(unix.CMSPAR)
// 	CRTSCTS  = uint32(unix.CRTSCTS)

// 	// linux-specific c_iflag bits:
// 	IUCLC = uint32(unix.IUCLC)
// 	IUTF8 = uint32(unix.IUTF8)

// 	// linux-specific c_oflag bits:
// 	OLCUC  = uint32(unix.OLCUC)
// 	OFILL  = uint32(unix.OFILL)
// 	OFDEL  = uint32(unix.OFDEL)
// 	NLDLY  = uint32(unix.NLDLY)
// 	NL0    = uint32(unix.NL0)
// 	NL1    = uint32(unix.NL1)
// 	CRDLY  = uint32(unix.CRDLY)
// 	CR0    = uint32(unix.CR0)
// 	CR1    = uint32(unix.CR1)
// 	CR2    = uint32(unix.CR2)
// 	CR3    = uint32(unix.CR3)
// 	TABDLY = uint32(unix.TABDLY)
// 	TAB0   = uint32(unix.TAB0)
// 	TAB1   = uint32(unix.TAB1)
// 	TAB2   = uint32(unix.TAB2)
// 	TAB3   = uint32(unix.TAB3)
// 	BSDLY  = uint32(unix.BSDLY)
// 	BS0    = uint32(unix.BS0)
// 	BS1    = uint32(unix.BS1)
// 	FFDLY  = uint32(unix.FFDLY)
// 	FF0    = uint32(unix.FF0)
// 	FF1    = uint32(unix.FF1)

// 	// linux-specific c_lflag bits:
// 	XCASE = uint32(unix.XCASE)
)

// Apparently missing from sys/unix
const (
	// taken from bits/termios.h
	TCOOFF = 0
	TCOON  = 1
	TCIOFF = 2
	TCION  = 3
)

const (
	getAttrIOCTL      = unix.TCGETS
	setAttrNowIOCTL   = unix.TCSETS
	setAttrDrainIOCTL = unix.TCSETSW
	setAttrFlushIOCTL = unix.TCSETSF
	flushIOCTL        = unix.TCFLSH
)

// GetLock gets the locking status of the termios structure of the given
// terminal. See tty_ioctl(4).
func GetLock(fd uintptr) (*Termios, error) {
	var tio Termios
	if err := ioctl(fd, unix.TIOCGLCKTRMIOS, unsafe.Pointer(&tio)); err != nil {
		return nil, err
	}

	return &tio, nil
}

// SetLock sets the locking status of the termios structure of the given
// terminal. Needs CAP_SYS_ADMIN. See tty_ioctl(4).
func (tio *Termios) SetLock(fd uintptr) error {
	return ioctl(fd, unix.TIOCSLCKTRMIOS, unsafe.Pointer(&tio))
}

// from the Bnnnnnn constants above
var speeds = [...]int{
	// first 16:
	0, 50, 75, 110, 134, 150, 200, 300, 600, 1200, 1800, 2400, 4800, 9600, 19200, 38400,
	// extended:
	57600, 115200, 230400, 460800, 500000, 576000, 921600, 1000000,
	1152000, 1500000, 2000000, 2500000, 3000000, 3500000, 4000000,
}

// GetSpeed gets the stored baud rate.
//
// NOTE linux doesn't support different speeds for in and out; everything is outSpeed
func (tio *Termios) GetSpeed() (in int, out int) {
	spID := int(tio.Cflag & CBAUD)
	if spID <= 0xf {
		return speeds[spID], speeds[spID]
	}
	spID -= int(CBAUDEX) - 0xf
	if 0xf >= spID || spID >= len(speeds) {
		return -1, -1
	}

	return speeds[spID], speeds[spID]
}

// SetSpeed sets the baud rate.
//
// NOTE linux doesn't support different speeds for in and out; everything is outSpeed.
func (tio *Termios) SetSpeed(_ int, speed int) error {
	spID := sort.SearchInts(speeds[:], speed)
	if spID >= len(speeds) || speed != speeds[spID] {
		return fmt.Errorf("%d is not a good rate, even for a baud rate", speed)
	}
	val := uint32(spID)
	if val > 0xf {
		val += CBAUDEX - 0xf
	}
	tio.Cflag = (tio.Cflag &^ CBAUD) | val

	return nil
}

const (
	rawImaskOff = ^uint32(unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON)
	rawImaskOn  = uint32(1<<32 - 1)
	rawOmaskOff = ^uint32(unix.OPOST)
	rawOmaskOn  = uint32(1<<32 - 1)
	rawLmaskOff = ^uint32(unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN)
	rawLmaskOn  = uint32(1<<32 - 1)
	rawCmaskOff = ^uint32(unix.CSIZE | unix.PARENB)
	rawCmaskOn  = uint32(unix.CS8)
)

// SendBreak transmits a continuous stream of zeros for between 0.25 and 0.5
// seconds if duration is 0, or for the given number of deciseconds if not, if
// the terminal supports breaks.
//
// Note this is TCSBRKP in tty_ioctl(4); this one seems saner than
// TCSBRK/tcsendbreak(3).
func SendBreak(fd uintptr, duration int) error {
	return ioctlu(fd, unix.TCSBRKP, uintptr(duration))
}

// Drain waits until all output written to the terminal referenced by fd has
// been transmitted to the terminal.
func Drain(fd uintptr) error {
	// on linux tcdrain is TCSBRK with non-zero arg
	return ioctlu(fd, unix.TCSBRK, uintptr(1))
}

// func (q Queue) bits() uintptr {
// 	switch q {
// 	case InputQueue:
// 		return unix.TCIFLUSH
// 	case OutputQueue:
// 		return unix.TCOFLUSH
// 	default:
// 		return unix.TCIOFLUSH
// 	}
// }

// Flow manages the suspending of data transmission or reception on the
// terminal referenced by fd.  The value of action must be one of the
// following:
//   TCOOFF  Suspend output.
//   TCOON   Restart suspended output.
//   TCIOFF  Transmit a STOP character (the XOFF in XON/XOFF)
//   TCION   Transmit a START character (the XON).
func Flow(fd uintptr, action int) error {
	return ioctlu(fd, unix.TCXONC, uintptr(action))
}

func setFlag(flag *uint32, mask uint32, v bool) {
	if v {
		*flag |= mask
	} else {
		*flag &= ^mask
	}
}

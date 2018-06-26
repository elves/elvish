package core

import (
	"fmt"
	"os"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/sys"
)

type TTY interface {
	Setuper
	Sizer
	Reader
	Writer
}

type Setuper interface {
	Setup() (restore func(), err error)
}

type Sizer interface {
	Size() (h, w int)
}

type Reader interface {
	StartRead() <-chan tty.Event
	SetRaw(raw bool)
	StopRead()
}

type Writer interface {
	Buffer() *ui.Buffer
	ResetBuffer()
	UpdateBuffer(bufNotes, bufMain *ui.Buffer, full bool) error
}

type aTTY struct {
	in, out *os.File
	r       tty.Reader
	w       tty.Writer
}

func newTTY(in, out *os.File) TTY {
	return &aTTY{in, out, nil, tty.NewWriter(out)}
}

func (t *aTTY) Setup() (func(), error) {
	restore, err := tty.Setup(t.in, t.out)
	return func() {
		err := restore()
		if err != nil {
			fmt.Println(t.out, "failed to restore terminal properties:", err)
		}
	}, err
}

func (t *aTTY) Size() (h, w int) {
	return sys.GetWinsize(t.out)
}

func (t *aTTY) StartRead() <-chan tty.Event {
	t.r = tty.NewReader(t.in)
	return t.r.EventChan()
}

func (t *aTTY) SetRaw(raw bool) {
	t.r.SetRaw(raw)
}

func (t *aTTY) StopRead() {
	t.r.Stop()
	t.r.Close()
	t.r = nil
}

func (t *aTTY) Buffer() *ui.Buffer {
	return t.w.CurrentBuffer()
}

func (t *aTTY) ResetBuffer() {
	t.w.ResetCurrentBuffer()
}

func (t *aTTY) UpdateBuffer(bufNotes, bufMain *ui.Buffer, full bool) error {
	return t.w.CommitBuffer(bufNotes, bufMain, full)
}

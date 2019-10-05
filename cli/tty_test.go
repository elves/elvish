package cli

import (
	"os"
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

func TestFakeTTY_Setup(t *testing.T) {
	tty, ttyCtrl := NewFakeTTY()
	restoreCalled := 0
	ttyCtrl.SetSetup(func() { restoreCalled++ }, nil)

	restore, err := tty.Setup()
	if err != nil {
		t.Errorf("Setup -> error %v, want nil", err)
	}
	restore()
	if restoreCalled != 1 {
		t.Errorf("Setup did not return restore")
	}
}

func TestFakeTTY_Size(t *testing.T) {
	tty, ttyCtrl := NewFakeTTY()
	ttyCtrl.SetSize(20, 30)
	h, w := tty.Size()
	if h != 20 || w != 30 {
		t.Errorf("Size -> (%v, %v), want (20, 30)", h, w)
	}
}

func TestFakeTTY_Events(t *testing.T) {
	tty, ttyCtrl := NewFakeTTY()
	events := tty.StartInput()
	ttyCtrl.Inject(term.K('a'), term.K('b'))
	event := <-events
	if event != term.K('a') {
		t.Errorf("Got event %v, want K('a')", event)
	}
	event = <-events
	if event != term.K('b') {
		t.Errorf("Got event %v, want K('b')", event)
	}
}

func TestFakeTTY_Signals(t *testing.T) {
	tty, ttyCtrl := NewFakeTTY()
	signals := tty.NotifySignals()
	ttyCtrl.InjectSignal(os.Interrupt, os.Kill)
	signal := <-signals
	if signal != os.Interrupt {
		t.Errorf("Got signal %v, want %v", signal, os.Interrupt)
	}
	signal = <-signals
	if signal != os.Kill {
		t.Errorf("Got signal %v, want %v", signal, os.Kill)
	}
}

func TestFakeTTY_Buffer(t *testing.T) {
	bufNotes1 := ui.NewBufferBuilder(10).WritePlain("notes 1").Buffer()
	buf1 := ui.NewBufferBuilder(10).WritePlain("buf 1").Buffer()
	bufNotes2 := ui.NewBufferBuilder(10).WritePlain("notes 2").Buffer()
	buf2 := ui.NewBufferBuilder(10).WritePlain("buf 2").Buffer()
	bufNotes3 := ui.NewBufferBuilder(10).WritePlain("notes 3").Buffer()
	buf3 := ui.NewBufferBuilder(10).WritePlain("buf 3").Buffer()

	tty, ttyCtrl := NewFakeTTY()

	if ttyCtrl.LastNotesBuffer() != nil {
		t.Errorf("LastNotesBuffer -> %v, want nil", ttyCtrl.LastNotesBuffer())
	}
	if ttyCtrl.LastBuffer() != nil {
		t.Errorf("LastBuffer -> %v, want nil", ttyCtrl.LastBuffer())
	}

	tty.UpdateBuffer(bufNotes1, buf1, true)
	if ttyCtrl.LastNotesBuffer() != bufNotes1 {
		t.Errorf("LastBuffer -> %v, want %v", ttyCtrl.LastNotesBuffer(), bufNotes1)
	}
	if ttyCtrl.LastBuffer() != buf1 {
		t.Errorf("LastBuffer -> %v, want %v", ttyCtrl.LastBuffer(), buf1)
	}
	if !ttyCtrl.VerifyBuffer(buf1) {
		t.Errorf("VerifyBuffer(buf1) -> false, want true")
	}
	if !ttyCtrl.VerifyNotesBuffer(bufNotes1) {
		t.Errorf("VerifyBuffer(bufNotes1) -> false, want true")
	}

	tty.UpdateBuffer(bufNotes2, buf2, true)
	if ttyCtrl.LastNotesBuffer() != bufNotes2 {
		t.Errorf("LastBuffer -> %v, want %v", ttyCtrl.LastNotesBuffer(), bufNotes2)
	}
	if ttyCtrl.LastBuffer() != buf2 {
		t.Errorf("LastBuffer -> %v, want %v", ttyCtrl.LastBuffer(), buf2)
	}
	if !ttyCtrl.VerifyBuffer(buf2) {
		t.Errorf("VerifyBuffer(buf2) -> false, want true")
	}
	if !ttyCtrl.VerifyNotesBuffer(bufNotes2) {
		t.Errorf("VerifyBuffer(bufNotes2) -> false, want true")
	}

	// Test Test{,Notes}Buffer
	tty.UpdateBuffer(bufNotes3, buf3, true)
	ttyCtrl.TestBuffer(t, buf3)
	ttyCtrl.TestNotesBuffer(t, bufNotes3)
	// Cannot test the failure branch as that will fail the test

	wantBufs := []*ui.Buffer{buf1, buf2, buf3}
	wantNotesBufs := []*ui.Buffer{bufNotes1, bufNotes2, bufNotes3}
	if !reflect.DeepEqual(ttyCtrl.BufferHistory(), wantBufs) {
		t.Errorf("BufferHistory did not return {buf1, buf2}")
	}
	if !reflect.DeepEqual(ttyCtrl.NotesBufferHistory(), wantNotesBufs) {
		t.Errorf("NotesBufferHistory did not return {bufNotes1, bufNotes2}")
	}
}

package clitest

import (
	"os"
	"reflect"
	"testing"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
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

func TestFakeTTY_SetRawInput(t *testing.T) {
	tty, ttyCtrl := NewFakeTTY()
	tty.SetRawInput(2)
	if raw := ttyCtrl.RawInput(); raw != 2 {
		t.Errorf("RawInput() -> %v, want 2", raw)
	}
}

func TestFakeTTY_Events(t *testing.T) {
	tty, ttyCtrl := NewFakeTTY()
	ttyCtrl.Inject(term.K('a'), term.K('b'))
	if event, err := tty.ReadEvent(); event != term.K('a') || err != nil {
		t.Errorf("Got (%v, %v), want (%v, nil)", event, err, term.K('a'))
	}
	if event := <-ttyCtrl.EventCh(); event != term.K('b') {
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
	bufNotes1 := term.NewBufferBuilder(10).Write("notes 1").Buffer()
	buf1 := term.NewBufferBuilder(10).Write("buf 1").Buffer()
	bufNotes2 := term.NewBufferBuilder(10).Write("notes 2").Buffer()
	buf2 := term.NewBufferBuilder(10).Write("buf 2").Buffer()
	bufNotes3 := term.NewBufferBuilder(10).Write("notes 3").Buffer()
	buf3 := term.NewBufferBuilder(10).Write("buf 3").Buffer()

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
	ttyCtrl.TestBuffer(t, buf1)
	ttyCtrl.TestNotesBuffer(t, bufNotes1)

	tty.UpdateBuffer(bufNotes2, buf2, true)
	if ttyCtrl.LastNotesBuffer() != bufNotes2 {
		t.Errorf("LastBuffer -> %v, want %v", ttyCtrl.LastNotesBuffer(), bufNotes2)
	}
	if ttyCtrl.LastBuffer() != buf2 {
		t.Errorf("LastBuffer -> %v, want %v", ttyCtrl.LastBuffer(), buf2)
	}
	ttyCtrl.TestBuffer(t, buf2)
	ttyCtrl.TestNotesBuffer(t, bufNotes2)

	// Test Test{,Notes}Buffer
	tty.UpdateBuffer(bufNotes3, buf3, true)
	ttyCtrl.TestBuffer(t, buf3)
	ttyCtrl.TestNotesBuffer(t, bufNotes3)
	// Cannot test the failure branch as that will fail the test

	wantBufs := []*term.Buffer{buf1, buf2, buf3}
	wantNotesBufs := []*term.Buffer{bufNotes1, bufNotes2, bufNotes3}
	if !reflect.DeepEqual(ttyCtrl.BufferHistory(), wantBufs) {
		t.Errorf("BufferHistory did not return {buf1, buf2}")
	}
	if !reflect.DeepEqual(ttyCtrl.NotesBufferHistory(), wantNotesBufs) {
		t.Errorf("NotesBufferHistory did not return {bufNotes1, bufNotes2}")
	}
}

func TestFakeTTY_ClearScreen(t *testing.T) {
	fakeTTY, ttyCtrl := NewFakeTTY()
	for i := 0; i < 5; i++ {
		if cleared := ttyCtrl.ScreenCleared(); cleared != i {
			t.Errorf("ScreenCleared -> %v, want %v", cleared, i)
		}
		fakeTTY.ClearScreen()
	}
}

func TestGetTTYCtrl_FakeTTY(t *testing.T) {
	fakeTTY, ttyCtrl := NewFakeTTY()
	if got, ok := GetTTYCtrl(fakeTTY); got != ttyCtrl || !ok {
		t.Errorf("-> %v, %v, want %v, %v", got, ok, ttyCtrl, true)
	}
}

func TestGetTTYCtrl_RealTTY(t *testing.T) {
	realTTY := cli.NewTTY(os.Stdin, os.Stderr)
	if _, ok := GetTTYCtrl(realTTY); ok {
		t.Errorf("-> _, true, want _, false")
	}
}

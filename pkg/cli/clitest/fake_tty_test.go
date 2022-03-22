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
		t.Errorf("FakeTTY Setup():\nWanted: %v\nActual: %v", nil, err)
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
		t.Errorf("FakeTTY Size():\nWanted: 20, 30\nActual: %v, %v", h, w)
	}
}

func TestFakeTTY_SetRawInput(t *testing.T) {
	tty, ttyCtrl := NewFakeTTY()
	tty.SetRawInput(2)
	if raw := ttyCtrl.RawInput(); raw != 2 {
		t.Errorf("FakeTTY RawInput():\nWanted: %v\nActual: %v", 2, raw)
	}
}

func TestFakeTTY_Events(t *testing.T) {
	tty, ttyCtrl := NewFakeTTY()
	ttyCtrl.Inject(term.K('a'), term.K('b'))
	if event, err := tty.ReadEvent(); event != term.K('a') || err != nil {
		t.Errorf("FakeTTY ReadEvent():\nWanted: %v, %v\nActual: %v, %v",
			term.K('a'), nil, event, err)
	}
	if event := <-ttyCtrl.EventCh(); event != term.K('b') {
		t.Errorf("FakeTTY event:\nWanted: %v\nActual: %v", term.K('b'), event)
	}
}

func TestFakeTTY_Signals(t *testing.T) {
	tty, ttyCtrl := NewFakeTTY()
	signals := tty.NotifySignals()
	ttyCtrl.InjectSignal(os.Interrupt, os.Kill)
	signal := <-signals
	if signal != os.Interrupt {
		t.Errorf("FakeTTY unexpected signal:\nWanted: %v\nActual: %v", os.Interrupt, signal)
	}
	signal = <-signals
	if signal != os.Kill {
		t.Errorf("FakeTTY unexpected signal:\nWanted: %v\nActual: %v", os.Kill, signal)
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
		t.Errorf("FakeTTY LastNotesBuffer():\nWanted: %v\nActual: %v",
			nil, ttyCtrl.LastNotesBuffer())
	}
	if ttyCtrl.LastBuffer() != nil {
		t.Errorf("FakeTTY LastBuffer():\nWanted: %v\nActual: %v",
			nil, ttyCtrl.LastBuffer())
	}

	tty.UpdateBuffer(bufNotes1, buf1, true)
	if ttyCtrl.LastNotesBuffer() != bufNotes1 {
		t.Errorf("FakeTTY LastNotesBuffer():\nWanted: %v\nActual: %v",
			bufNotes1, ttyCtrl.LastNotesBuffer())
	}
	if ttyCtrl.LastBuffer() != buf1 {
		t.Errorf("FakeTTY LastBuffer():\nWanted: %v\nActual: %v",
			buf1, ttyCtrl.LastBuffer())
	}
	ttyCtrl.TestBuffer(t, buf1)
	ttyCtrl.TestNotesBuffer(t, bufNotes1)

	tty.UpdateBuffer(bufNotes2, buf2, true)
	if ttyCtrl.LastNotesBuffer() != bufNotes2 {
		t.Errorf("FakeTTY LastNotesBuffer():\nWanted: %v\nActual: %v",
			bufNotes2, ttyCtrl.LastNotesBuffer())
	}
	if ttyCtrl.LastBuffer() != buf2 {
		t.Errorf("FakeTTY LastBuffer():\nWanted: %v\nActual: %v",
			buf2, ttyCtrl.LastBuffer())
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
			t.Errorf("FakeTTY ScreenCleared():\nWanted: %v\nActual: %v", i, cleared)
		}
		fakeTTY.ClearScreen()
	}
}

func TestGetTTYCtrl_FakeTTY(t *testing.T) {
	fakeTTY, ttyCtrl := NewFakeTTY()
	if got, ok := GetTTYCtrl(fakeTTY); got != ttyCtrl || !ok {
		t.Errorf("GetTTYCtrl():\nWanted: %v, %v\nActual: %v, %v", ttyCtrl, true, got, ok)
	}
}

func TestGetTTYCtrl_RealTTY(t *testing.T) {
	realTTY := cli.NewTTY(os.Stdin, os.Stderr)
	if _, ok := GetTTYCtrl(realTTY); ok {
		t.Errorf("GetTTYCtrl() error:\nWanted: %v\nActual: %v", false, true)
	}
}

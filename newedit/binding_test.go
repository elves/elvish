package newedit

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/eval"
)

func TestCallBinding_CallsFunction(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	called := 0
	callBinding(nt, ev, eval.NewBuiltinFn("test binding", func(fm *eval.Frame) {
		called++
	}))
	if called != 1 {
		t.Errorf("binding called %v times, want once", called)
	}
}

func TestCallBinding_NotifyOnValueOutput(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	callBinding(nt, ev, eval.NewBuiltinFn("test binding", func(fm *eval.Frame) {
		fm.OutputChan() <- "VALUE"
	}))
	wantNotes := []string{"[value out] VALUE"}
	if !reflect.DeepEqual(nt.notes, wantNotes) {
		t.Errorf("got notes %v, want %v", nt.notes, wantNotes)
	}
}

func TestCallBinding_NotifyOnByteOutput(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	callBinding(nt, ev, eval.NewBuiltinFn("test binding", func(fm *eval.Frame) {
		fm.OutputFile().WriteString("BYTES")
	}))
	wantNotes := []string{"[bytes out] BYTES"}
	if !reflect.DeepEqual(nt.notes, wantNotes) {
		t.Errorf("got notes %v, want %v", nt.notes, wantNotes)
	}
}

func TestCallBinding_StripsNewlinesFromByteOutput(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	callBinding(nt, ev, eval.NewBuiltinFn("test binding", func(fm *eval.Frame) {
		fm.OutputFile().WriteString("line 1\nline 2\n")
	}))
	wantNotes := []string{"[bytes out] line 1", "[bytes out] line 2"}
	if !reflect.DeepEqual(nt.notes, wantNotes) {
		t.Errorf("got notes %v, want %v", nt.notes, wantNotes)
	}
}

func TestCallBinding_NotifyOnError(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	callBinding(nt, ev, eval.NewBuiltinFn("test binding", func() error {
		return errors.New("ERROR")
	}))
	wantNotes := []string{"[binding error] ERROR"}
	if !reflect.DeepEqual(nt.notes, wantNotes) {
		t.Errorf("got notes %v, want %v", nt.notes, wantNotes)
	}
}

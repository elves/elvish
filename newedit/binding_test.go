package newedit

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/newedit/utils"
)

func TestKeyHandlerFromBinding(t *testing.T) {
	called := 0
	binding := buildBinding(
		"a", eval.NewBuiltinFn("test binding", func() { called++ }))
	handler := keyHandlerFromBinding(dummyNotifier{}, eval.NewEvaler(), &binding)

	action := handler(ui.Key{Rune: 'a'})

	if called != 1 {
		t.Errorf("Binding called %d, want once", called)
	}
	if action != types.NoAction {
		t.Errorf("Binding returned %v, want NoAction", action)
	}
}

func TestKeyHandlerFromBinding_Fallback(t *testing.T) {
	nt := &fakeNotifier{}
	binding := EmptyBindingMap
	handler := keyHandlerFromBinding(nt, eval.NewEvaler(), &binding)

	action := handler(ui.Key{Rune: 'a'})

	wantNotes := []string{"Unbound: a"}
	if !reflect.DeepEqual(nt.notes, wantNotes) {
		t.Errorf("Notes %v, want %v", nt.notes, wantNotes)
	}
	if action != types.NoAction {
		t.Errorf("Fallback binding returned %v, want NoAction", action)
	}
}

func buildBinding(data ...interface{}) BindingMap {
	binding := EmptyBindingMap
	for i := 0; i < len(data); i += 2 {
		result, err := EmptyBindingMap.Assoc(data[i], data[i+1])
		if err != nil {
			panic(err)
		}
		binding = result.(BindingMap)
	}
	return binding
}

func TestCallBinding_CallsFunction(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	called := 0
	callBinding(nt, ev, eval.NewBuiltinFn("test binding", func() {
		called++
	}))
	if called != 1 {
		t.Errorf("binding called %v times, want once", called)
	}
}

func TestCallBinding_CapturesAction(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	action := callBinding(nt, ev, eval.NewBuiltinFn("test", func() error {
		return utils.ActionError(types.CommitCode)
	}))
	if action != types.CommitCode {
		t.Errorf("got ret = %v, want %v", action, types.CommitCode)
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

package newedit

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/clitypes"
	"github.com/elves/elvish/newedit/cliutil"
)

func TestKeyHandlerFromBinding_CallsBinding(t *testing.T) {
	called := 0
	binding := buildBinding(
		"a", eval.NewGoFn("[test]", func() { called++ }))
	handler := keyHandlerFromBindings(&fakeEditor{}, eval.NewEvaler(), &binding)

	action := handler(ui.Key{Rune: 'a'})

	if called != 1 {
		t.Errorf("Binding called %d, want once", called)
	}
	if action != clitypes.NoAction {
		t.Errorf("Binding returned %v, want NoAction", action)
	}
}

func TestKeyHandlerFromBinding_SetsBindingKey(t *testing.T) {
	ed := &fakeEditor{}
	var gotKey ui.Key
	binding := buildBinding(
		"a", eval.NewGoFn("[test]", func() { gotKey = ed.State().BindingKey() }))
	handler := keyHandlerFromBindings(ed, eval.NewEvaler(), &binding)

	key := ui.Key{Rune: 'a'}
	_ = handler(key)

	if gotKey != key {
		t.Errorf("Got key %v, want %v", gotKey, key)
	}
}

func TestKeyHandlerFromBinding_Unbound(t *testing.T) {
	ed := &fakeEditor{}
	binding := emptyBindingMap
	handler := keyHandlerFromBindings(ed, eval.NewEvaler(), &binding)

	action := handler(ui.Key{Rune: 'a'})

	wantNotes := []string{"Unbound: a"}
	if !reflect.DeepEqual(ed.fakeNotifier.notes, wantNotes) {
		t.Errorf("Notes %v, want %v", ed.fakeNotifier.notes, wantNotes)
	}
	if action != clitypes.NoAction {
		t.Errorf("Fallback binding returned %v, want NoAction", action)
	}
}

func TestIndexLayeredBindings(t *testing.T) {
	// In a two-layer binding, there are 4 different cases:
	// 1. Key is bound in upper layer
	// 2. Key is bound not in upper layer but in lower layer
	// 3. Key
	called := 0
	binding1 := buildBinding(
		"a", eval.NewGoFn("[a1]", func() { called = 1 }))
	binding2 := buildBinding(
		"a", eval.NewGoFn("[a2]", func() { called = 2 }),
		"b", eval.NewGoFn("[b2]", func() { called = 2 }))

	handler := keyHandlerFromBindings(&fakeEditor{}, eval.NewEvaler(),
		&binding1, &binding2)

	// Prefer upper layer when present in both
	_ = handler(ui.K('a'))
	if called != 1 {
		t.Errorf("want a1 to be called, got %d", called)
	}

	// Use lower layer when absent in upper layer
	called = 0
	_ = handler(ui.K('b'))
	if called != 2 {
		t.Errorf("want b2 to be called, got %d", called)
	}

	// Use lower layer default when upper layer does not have default
	b, _ := binding2.Assoc(
		ui.Default, eval.NewGoFn("[d2]", func() { called = 2 }))
	binding2 = b.(bindingMap)

	called = 0
	_ = handler(ui.K('d'))
	if called != 2 {
		t.Errorf("want d2 to be called, got %d", called)
	}

	// Prefer upper layer default
	b, _ = binding1.Assoc(
		ui.Default, eval.NewGoFn("[d1]", func() { called = 1 }))
	binding1 = b.(bindingMap)

	called = 0
	_ = handler(ui.K('d'))
	if called != 1 {
		t.Errorf("want d1 to be called, got %d", called)
	}

	// Exact matches in all layers are tried before falling back to default
	b, _ = binding2.Assoc(
		"c", eval.NewGoFn("[c2]", func() { called = 2 }))
	binding2 = b.(bindingMap)

	called = 0
	_ = handler(ui.K('c'))
	if called != 2 {
		t.Errorf("want c2 to be called, got %d", called)
	}
}

func buildBinding(data ...interface{}) bindingMap {
	binding := emptyBindingMap
	for i := 0; i < len(data); i += 2 {
		result, err := emptyBindingMap.Assoc(data[i], data[i+1])
		if err != nil {
			panic(err)
		}
		binding = result.(bindingMap)
	}
	return binding
}

func TestCallBinding_CallsFunction(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	called := 0
	callBinding(nt, ev, eval.NewGoFn("test binding", func() {
		called++
	}))
	if called != 1 {
		t.Errorf("binding called %v times, want once", called)
	}
}

func TestCallBinding_CapturesAction(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	action := callBinding(nt, ev, eval.NewGoFn("test", func() error {
		return cliutil.ActionError(clitypes.CommitCode)
	}))
	if action != clitypes.CommitCode {
		t.Errorf("got ret = %v, want %v", action, clitypes.CommitCode)
	}
}

func TestCallBinding_NotifyOnValueOutput(t *testing.T) {
	ev := eval.NewEvaler()
	nt := &fakeNotifier{}

	callBinding(nt, ev, eval.NewGoFn("test binding", func(fm *eval.Frame) {
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

	callBinding(nt, ev, eval.NewGoFn("test binding", func(fm *eval.Frame) {
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

	callBinding(nt, ev, eval.NewGoFn("test binding", func(fm *eval.Frame) {
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

	callBinding(nt, ev, eval.NewGoFn("test binding", func() error {
		return errors.New("ERROR")
	}))
	wantNotes := []string{"[binding error] ERROR"}
	if !reflect.DeepEqual(nt.notes, wantNotes) {
		t.Errorf("got notes %v, want %v", nt.notes, wantNotes)
	}
}

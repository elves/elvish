package newedit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
)

var abbrData = [][2]string{{"xx", "xx full"}, {"yy", "yy full"}}

func TestInitInsert_Abbr(t *testing.T) {
	m, ns := initInsert(&fakeApp{}, eval.NewEvaler())

	abbrValue := vals.EmptyMap
	for _, pair := range abbrData {
		abbrValue = abbrValue.Assoc(pair[0], pair[1])
	}
	ns["abbr"].Set(abbrValue)

	var cbData [][2]string
	m.AbbrIterate(func(a, f string) {
		cbData = append(cbData, [2]string{a, f})
	})

	if !reflect.DeepEqual(cbData, abbrData) {
		t.Errorf("Callback called with %v, want %v", cbData, abbrData)
	}
}

func TestInitInsert_Binding(t *testing.T) {
	m, ns := initInsert(&fakeApp{}, eval.NewEvaler())
	called := 0
	binding, err := emptyBindingMap.Assoc("a",
		eval.NewGoFn("test binding", func() { called++ }))
	if err != nil {
		panic(err)
	}
	ns["binding"].Set(binding)

	m.HandleEvent(term.KeyEvent{Rune: 'a'}, &clitypes.State{})

	if called != 1 {
		t.Errorf("Handler called %d times, want once", called)
	}
}

func TestInitInsert_QuotePaste(t *testing.T) {
	m, ns := initInsert(&fakeApp{}, eval.NewEvaler())

	ns["quote-paste"].Set(true)

	if !m.Config.QuotePaste() {
		t.Errorf("QuotePaste not set via namespae")
	}
}

func TestInitInsert_Start(t *testing.T) {
	ed := &fakeApp{}
	ev := eval.NewEvaler()
	m, ns := initInsert(ed, ev)

	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[test]"), nil)
	fm.Call(getFn(ns, "start"), nil, eval.NoOpts)

	if ed.state.Mode() != m {
		t.Errorf("state is not insert mode after calling start")
	}
}

func TestInitInsert_DefaultHandler(t *testing.T) {
	ed := &fakeApp{}
	ev := eval.NewEvaler()
	_, ns := initInsert(ed, ev)

	// Pretend that we are executing a binding for "a".
	ed.state.SetBindingKey(ui.Key{Rune: 'a'})

	// Call <edit:insert>:default-binding.
	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[test]"), nil)
	fm.Call(getFn(ns, "default-handler"), nil, eval.NoOpts)

	// Verify that the default handler has executed, inserting "a".
	if ed.state.Raw.Code != "a" {
		t.Errorf("state.Raw.Code = %q, want %q", ed.state.Raw.Code, "a")
	}
}

func getFn(ns eval.Ns, name string) eval.Callable {
	return ns[name+eval.FnSuffix].Get().(eval.Callable)
}

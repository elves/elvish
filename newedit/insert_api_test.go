package newedit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/newedit/types"
)

var abbrData = [][2]string{{"xx", "xx full"}, {"yy", "yy full"}}

func TestInitInsert_Abbr(t *testing.T) {
	m, ns := initInsert(dummyEditor{}, eval.NewEvaler())

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
	m, ns := initInsert(dummyEditor{}, eval.NewEvaler())
	called := 0
	binding, err := EmptyBindingMap.Assoc("a",
		eval.NewBuiltinFn("test binding", func() { called++ }))
	if err != nil {
		panic(err)
	}
	ns["binding"].Set(binding)

	m.HandleEvent(tty.KeyEvent{Rune: 'a'}, &types.State{})

	if called != 1 {
		t.Errorf("Handler called %d times, want once", called)
	}
}

func TestInitInsert_QuotePaste(t *testing.T) {
	m, ns := initInsert(dummyEditor{}, eval.NewEvaler())

	ns["quote-paste"].Set(true)

	if !m.Config.QuotePaste() {
		t.Errorf("QuotePaste not set via namespae")
	}
}

func TestInitInsert_Start(t *testing.T) {
	ed := &fakeEditor{}
	ev := eval.NewEvaler()
	m, ns := initInsert(ed, ev)

	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[test]"), nil)
	fm.Call(ns["start"+eval.FnSuffix].Get().(eval.Callable), nil, eval.NoOpts)

	if ed.state.Mode() != m {
		t.Errorf("state is not insert mode after calling start")
	}
}

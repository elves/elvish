package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
)

func setupAPI() (*cli.App, *eval.Evaler, eval.Ns) {
	app := cli.NewApp(cli.AppSpec{TTY: cli.NewStdTTY()})
	ev := eval.NewEvaler()
	ns := eval.Ns{}
	initAPI(app, ev, ns)
	return app, ev, ns
}

func TestInitAPI_BeforeReadline(t *testing.T) {
	app, _, ns := setupAPI()

	var called int
	ns["before-readline"].Set(vals.MakeList(eval.NewGoFn("[test]", func() {
		called++
	})))
	app.AppSpec.BeforeReadline()
	if called != 1 {
		t.Errorf("before-readline called %d times, want once", called)
	}
}

func TestInitAPI_AfterReadline(t *testing.T) {
	app, _, ns := setupAPI()

	var called int
	var calledWith string
	ns["after-readline"].Set(vals.MakeList(eval.NewGoFn("[test]", func(s string) {
		called++
		calledWith = s
	})))
	app.AppSpec.AfterReadline("code")
	if called != 1 {
		t.Errorf("after-readline called %d times, want once", called)
	}
	if calledWith != "code" {
		t.Errorf("after-readline called with %q, want %q", calledWith, "code")
	}
}

/*
func TestInitAPI_Insert_Abbr(t *testing.T) {
	app, _, ns := setupAPI()
	m := vals.MakeMap("xx", "xx full", "yy", "yy full")
	getNs(ns, "insert")["abbr"].Set(m)

	collected := vals.EmptyMap
	app.CodeArea.Abbreviations(func(a, f string) {
		collected = collected.Assoc(a, f)
	})

	if !vals.Equal(m, collected) {
		t.Errorf("Callback collected %v, var set %v", collected, m)
	}
}

func TestInitAPI_Insert_Binding(t *testing.T) {
	app, _, ns := setupAPI()
	testKeyBinding(t, getNs(ns, "insert")["binding"], app.CodeArea.OverlayHandler)
}

func TestInitAPI_Insert_QuotePaste(t *testing.T) {
	app, _, ns := setupAPI()
	for _, quote := range []bool{false, true} {
		getNs(ns, "insert")["quote-paste"].Set(quote)
		if got := app.CodeArea.QuotePaste(); got != quote {
			t.Errorf("quote paste = %v, want %v", got, quote)
		}
	}
}
*/

func testKeyBinding(t *testing.T, v vars.Var, h el.Handler) {
	t.Helper()

	var called int
	binding, err := emptyBindingMap.Assoc(
		"a", eval.NewGoFn("[binding]", func() { called++ }))
	if err != nil {
		panic(err)
	}
	v.Set(binding)

	handled := h.Handle(term.K('a'))

	if !handled {
		t.Errorf("handled = false, want true")
	}
	if called != 1 {
		t.Errorf("handler called %d times, want once", called)
	}
}

func getNs(ns eval.Ns, name string) eval.Ns {
	return ns[name+eval.NsSuffix].Get().(eval.Ns)
}

func getFn(ns eval.Ns, name string) eval.Callable {
	return ns[name+eval.FnSuffix].Get().(eval.Callable)
}

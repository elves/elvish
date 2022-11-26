package edit

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/ui"
)

func TestBindingMap(t *testing.T) {
	// TODO
	TestWithSetup(t, func(ev *eval.Evaler) {
		ev.ExtendBuiltin(eval.BuildNs().AddGoFn("binding-map", makeBindingMap))
	},
		// Checking key and value when constructing
		That("binding-map [&[]={ }]").
			Throws(ErrorWithMessage("must be key or string")),
		That("binding-map [&foo={ }]").
			Throws(ErrorWithMessage("bad key: foo")),
		That("binding-map [&a=string]").
			Throws(ErrorWithMessage("value should be function")),

		// repr prints a binding-map like an ordinary map
		That("repr (binding-map [&])").Prints("[&]\n"),
		// Keys are always sorted
		That("repr (binding-map [&a=$nop~ &b=$nop~ &c=$nop~])").
			Prints("[&a=<builtin nop> &b=<builtin nop> &c=<builtin nop>]\n"),

		// Indexing
		That("eq $nop~ (binding-map [&a=$nop~])[a]").Puts(true),
		// Checking key
		That("put (binding-map [&a=$nop~])[foo]").
			Throws(ErrorWithMessage("bad key: foo")),

		// Assoc
		That("count (assoc (binding-map [&a=$nop~]) b $nop~)").Puts(2),
		// Checking key
		That("(assoc (binding-map [&a=$nop~]) foo $nop~)").
			Throws(ErrorWithMessage("bad key: foo")),
		// Checking value
		That("(assoc (binding-map [&a=$nop~]) b foo)").
			Throws(ErrorWithMessage("value should be function")),

		// Dissoc
		That("count (dissoc (binding-map [&a=$nop~]) a)").Puts(0),
		// Allows bad key - no op
		That("count (dissoc (binding-map [&a=$nop~]) foo)").Puts(1),
	)
}

func TestBindingHelp(t *testing.T) {
	a := func() {}
	b := func() {}
	c := func() {}
	ns := eval.BuildNs().AddGoFns(map[string]any{"a": a, "b": b, "c": c}).Ns()
	fn := func(name string) any { return ns.IndexString(name + eval.FnSuffix).Get() }

	entries := []bindingHelpEntry{{"do a", "a"}, {"do b", "b"}}

	// A bindings map with no relevant binding
	m0 := bindingsMap{vals.MakeMap(ui.K('C', ui.Ctrl), fn("c"))}
	want0 := ui.T("")
	if got := bindingHelp(m0, ns, entries...); !equalText(got, want0) {
		t.Errorf("got %v, want %v", got, want0)
	}

	// A map with one relevant binding
	m1 := bindingsMap{vals.MakeMap(ui.K('A', ui.Ctrl), fn("a"))}
	want1 := ui.MarkLines(
		"Ctrl-A do a", Styles,
		"++++++     ")
	if got := bindingHelp(m1, ns, entries...); !equalText(got, want1) {
		t.Errorf("got %v, want %v", got, want1)
	}

	// A map with bindings for both $a~ and $b~
	m2 := bindingsMap{vals.MakeMap(
		ui.K('A', ui.Ctrl), fn("a"), ui.K('B', ui.Ctrl), fn("b"))}
	want2 := ui.MarkLines(
		"Ctrl-A do a Ctrl-B do b", Styles,
		"++++++      ++++++     ")
	if got := bindingHelp(m2, ns, entries...); !equalText(got, want2) {
		t.Errorf("got %v, want %v", got, want2)
	}
}

func equalText(a, b ui.Text) bool {
	return reflect.DeepEqual(ui.NormalizeText(a), ui.NormalizeText(b))
}

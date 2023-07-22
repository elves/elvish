package edit

import (
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

func TestBindingMap(t *testing.T) {
	// TODO
	TestWithEvalerSetup(t, func(ev *eval.Evaler) {
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

// The happy path of bindingHelp is tested in modes that use bindingHelp.

func TestBindingHelp_NoBinding(t *testing.T) {
	ns := eval.BuildNs().
		AddGoFn("a", func() {}).
		AddVar("binding", vars.FromInit(bindingsMap{vals.EmptyMap})).
		Ns()

	// A bindings map with no relevant binding
	if got := bindingTips(ns, "binding", bindingTip("do a", "a")); len(got) > 0 {
		t.Errorf("got %v, want empty text", got)
	}
}

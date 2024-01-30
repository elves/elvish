package edit

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

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

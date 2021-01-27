package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vars"
)

func TestNs(t *testing.T) {
	Test(t,
		That("kind-of (ns [&])").Puts("ns"),
		// A Ns is only equal to itself
		That("ns = (ns [&]); eq $ns $ns").Puts(true),
		That("eq (ns [&]) (ns [&])").Puts(false),
		That("eq (ns [&]) [&]").Puts(false),

		That(`ns: = (ns [&a=b &x=y]); put $ns:a`).Puts("b"),
		That(`ns: = (ns [&a=b &x=y]); put $ns:[a]`).Puts("b"),
		// Test multi-key ns when sorting is possible
		That(`keys (ns [&a=b])`).Puts("a"),
		That(`has-key (ns [&a=b &x=y]) a`).Puts(true),
		That(`has-key (ns [&a=b &x=y]) b`).Puts(false),
	)
}

func TestBuiltinFunctionsReadOnly(t *testing.T) {
	Test(t,
		That("return~ = { }").Throws(vars.ErrSetReadOnlyVar, "return~ = { }"),
	)
}

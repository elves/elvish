package eval

import "testing"

func TestNs(t *testing.T) {
	runTests(t, []Test{
		NewTest("kind-of (ns)").WantOutStrings("ns"),
		// A Ns is only equal to itself
		NewTest("ns = (ns); eq $ns $ns").WantOutBools(true),
		NewTest("eq (ns) (ns)").WantOutBools(false),
		NewTest("eq (ns) [&]").WantOutBools(false),
	})
}

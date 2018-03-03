package eval

import "testing"

func TestNs(t *testing.T) {
	runTests(t, []Test{
		That("kind-of (ns [&])").Puts("ns"),
		// A Ns is only equal to itself
		That("ns = (ns [&]); eq $ns $ns").Puts(true),
		That("eq (ns [&]) (ns [&])").Puts(false),
		That("eq (ns [&]) [&]").Puts(false),
	})
}

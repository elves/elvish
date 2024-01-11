package vals

import (
	"fmt"
	"math/big"
	"os"
	"testing"

	"src.elv.sh/pkg/tt"
)

type reprer struct{}

func (reprer) Repr(int) string { return "<reprer>" }

type nonReprer struct{}

func TestReprPlain(t *testing.T) {
	tt.Test(t, ReprPlain,
		Args(nil).Rets("$nil"),

		Args(false).Rets("$false"),
		Args(true).Rets("$true"),

		Args("foo").Rets("foo"),

		Args(1).Rets("(num 1)"),
		Args(bigInt(z)).Rets("(num "+z+")"),
		Args(big.NewRat(1, 2)).Rets("(num 1/2)"),
		Args(1.0).Rets("(num 1.0)"),
		Args(1e10).Rets("(num 10000000000.0)"),

		Args(os.Stdin).Rets(
			fmt.Sprintf("<file{%s %d}>", os.Stdin.Name(), os.Stdin.Fd())),

		Args(EmptyList).Rets("[]"),
		Args(MakeList("foo", "bar")).Rets("[foo bar]"),

		Args(EmptyMap).Rets("[&]"),
		Args(MakeMap("foo", "bar")).Rets("[&foo=bar]"),
		// Keys of the same type are sorted.
		Args(MakeMap("b", "second", "a", "first", "c", "third")).
			Rets("[&a=first &b=second &c=third]"),
		Args(MakeMap(2, "second", 1, "first", 3, "third")).
			Rets("[&(num 1)=first &(num 2)=second &(num 3)=third]"),
		// Keys of mixed types tested in a different test.

		Args(reprer{}).Rets("<reprer>"),
		Args(nonReprer{}).Rets("<unknown {}>"),
	)
}

func TestReprPlain_MapWithKeysOfMixedTypes(t *testing.T) {
	m := MakeMap(
		"b", "second", "a", "first", "c", "third",
		2, "second", 1, "first", 3, "third")
	strPart := "&a=first &b=second &c=third"
	numPart := "&(num 1)=first &(num 2)=second &(num 3)=third"
	want1 := "[" + strPart + " " + numPart + "]"
	want2 := "[" + numPart + " " + strPart + "]"
	got := ReprPlain(m)
	if got != want1 && got != want2 {
		t.Errorf("got %q, want %q or %q", got, want1, want2)
	}
}

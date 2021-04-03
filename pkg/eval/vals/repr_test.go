package vals

import (
	"fmt"
	"os"
	"testing"

	. "src.elv.sh/pkg/tt"
)

type reprer struct{}

func (reprer) Repr(int) string { return "<reprer>" }

type nonReprer struct{}

func repr(a interface{}) string { return Repr(a, NoPretty) }

func TestRepr(t *testing.T) {
	Test(t, Fn("repr", repr), Table{
		Args(nil).Rets("$nil"),
		Args(false).Rets("$false"),
		Args(true).Rets("$true"),
		Args("foo").Rets("foo"),
		Args(1.0).Rets("(num 1.0)"),
		Args(1e10).Rets("(num 10000000000.0)"),
		Args(os.Stdin).Rets(
			fmt.Sprintf("<file{%s %d}>", os.Stdin.Name(), os.Stdin.Fd())),
		Args(EmptyList).Rets("[]"),
		Args(MakeList("foo", "bar")).Rets("[foo bar]"),
		Args(EmptyMap).Rets("[&]"),
		Args(MakeMap("foo", "bar")).Rets("[&foo=bar]"),
		Args(reprer{}).Rets("<reprer>"),
		Args(nonReprer{}).Rets("<unknown {}>"),
	})
}

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
	tt.Test(t, tt.Fn("ReprPlain", ReprPlain), tt.Table{
		tt.Args(nil).Rets("$nil"),

		tt.Args(false).Rets("$false"),
		tt.Args(true).Rets("$true"),

		tt.Args("foo").Rets("foo"),

		tt.Args(1).Rets("(num 1)"),
		tt.Args(bigInt(z)).Rets("(num " + z + ")"),
		tt.Args(big.NewRat(1, 2)).Rets("(num 1/2)"),
		tt.Args(1.0).Rets("(num 1.0)"),
		tt.Args(1e10).Rets("(num 10000000000.0)"),

		tt.Args(os.Stdin).Rets(
			fmt.Sprintf("<file{%s %d}>", os.Stdin.Name(), os.Stdin.Fd())),

		tt.Args(EmptyList).Rets("[]"),
		tt.Args(MakeList("foo", "bar")).Rets("[foo bar]"),

		tt.Args(EmptyMap).Rets("[&]"),
		tt.Args(MakeMap("foo", "bar")).Rets("[&foo=bar]"),

		tt.Args(reprer{}).Rets("<reprer>"),
		tt.Args(nonReprer{}).Rets("<unknown {}>"),
	})
}

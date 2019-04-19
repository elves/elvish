package vals

import (
	"fmt"
	"os"
	"testing"

	"github.com/elves/elvish/tt"
)

type reprer struct{}

func (reprer) Repr(int) string { return "<reprer>" }

type nonReprer struct{}

func repr(a interface{}) string { return Repr(a, NoPretty) }

func TestRepr(t *testing.T) {
	tt.Test(t, tt.Fn("repr", repr), tt.Table{
		Args(nil).Rets("$nil"),
		Args(false).Rets("$false"),
		Args(true).Rets("$true"),
		Args("foo").Rets("foo"),
		Args(1.0).Rets("(float64 1)"),
		Args(os.Stdin).Rets(
			fmt.Sprintf("<file{%s %d}>", os.Stdin.Name(), os.Stdin.Fd())),
		Args(EmptyList).Rets("[]"),
		Args(MakeList("foo", "bar")).Rets("[foo bar]"),
		Args(EmptyMap).Rets("[&]"),
		Args(MakeMap("foo", "bar")).Rets("[&foo=bar]"),
		Args(testStructMap{"name", 1.0}).Rets(
			"[&name=name &score-number=(float64 1)]"),
		Args(reprer{}).Rets("<reprer>"),
		Args(nonReprer{}).Rets("<unknown {}>"),
	})
}

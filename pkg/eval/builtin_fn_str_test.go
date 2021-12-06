package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
)

func TestStringComparisonCommands(t *testing.T) {
	Test(t,
		That(`<s a b`).Puts(true),
		That(`<s 2 10`).Puts(false),
		That(`<=s a a`).Puts(true),
		That(`<=s a b`).Puts(true),
		That(`<=s b a`).Puts(false),
		That(`==s haha haha`).Puts(true),
		That(`==s 10 10.0`).Puts(false),
		That(`!=s haha haha`).Puts(false),
		That(`!=s 10 10.1`).Puts(true),
		That(`>s a b`).Puts(false),
		That(`>s 2 10`).Puts(true),
		That(`>=s a a`).Puts(true),
		That(`>=s a b`).Puts(false),
		That(`>=s b a`).Puts(true),
	)
}

func TestToString(t *testing.T) {
	Test(t,
		That(`to-string str (num 1) $true`).Puts("str", "1", "$true"),
		thatOutputErrorIsBubbled("to-string str"),
	)
}

func TestBase(t *testing.T) {
	Test(t,
		That(`base 2 1 3 4 16 255`).Puts("1", "11", "100", "10000", "11111111"),
		That(`base 16 42 233`).Puts("2a", "e9"),
		That(`base 1 1`).Throws(ErrBadBase),
		That(`base 37 10`).Throws(ErrBadBase),
		thatOutputErrorIsBubbled("base 2 1"),
	)
}

func TestWcswidth(t *testing.T) {
	Test(t,
		That(`wcswidth 你好`).Puts(4),
		That(`-override-wcwidth x 10; wcswidth 1x2x; -override-wcwidth x 1`).
			Puts(22),
	)
}

func TestEawk(t *testing.T) {
	Test(t,
		That(`echo "  ax  by cz  \n11\t22 33" | eawk {|@a| put $a[-1] }`).
			Puts("cz", "33"),
		// Bad input type
		That(`num 42 | eawk {|@a| fail "this should not run" }`).
			Throws(ErrInputOfEawkMustBeString),
		// Propagation of exception
		That(`
			to-lines [1 2 3 4] | eawk {|@a|
				if (==s 3 $a[1]) {
					fail "stop eawk"
				}
				put $a[1]
			}
		`).Puts("1", "2").Throws(FailError{"stop eawk"}),
		// break
		That(`
			to-lines [" a" "b\tc " "d" "e"] | eawk {|@a|
				if (==s d $a[1]) {
					break
				} else {
					put $a[-1]
				}
			}
		`).Puts("a", "c"),
		// continue
		That(`
			to-lines [" a" "b\tc " "d" "e"] | eawk {|@a|
				if (==s d $a[1]) {
					continue
				} else {
					put $a[-1]
				}
			}
		`).Puts("a", "c", "e"),
	)
}

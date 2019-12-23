package eval

import (
	"testing"
)

func TestBool(t *testing.T) {
	Test(t,
		That(`bool $true`).Puts(true),
		That(`bool a`).Puts(true),
		That(`bool [a]`).Puts(true),
		// "Empty" values are also true in Elvish
		That(`bool []`).Puts(true),
		That(`bool [&]`).Puts(true),
		That(`bool 0`).Puts(true),
		That(`bool ""`).Puts(true),
		// Only errors and $false are false
		That(`bool ?(fail x)`).Puts(false),
		That(`bool $false`).Puts(false),

		That(`not $false`).Puts(true),
		That(`not ?(fail x)`).Puts(true),
		That(`not $true`).Puts(false),
		That(`not 0`).Puts(false),

		That(`is 1 1`).Puts(true),
		That(`is a b`).Puts(false),
		That(`is [] []`).Puts(true),
		That(`is [1] [1]`).Puts(false),

		That(`eq 1 1`).Puts(true),
		That(`eq a b`).Puts(false),
		That(`eq [] []`).Puts(true),
		That(`eq [1] [1]`).Puts(true),
		That(`eq 1 1 2`).Puts(false),

		That(`not-eq a b`).Puts(true),
		That(`not-eq a a`).Puts(false),
		// not-eq is true as long as each adjancent pair is not equal.
		That(`not-eq 1 2 1`).Puts(true),
	)
}

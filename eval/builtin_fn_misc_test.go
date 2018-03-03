package eval

import (
	"testing"
)

func TestBuiltinFnMisc(t *testing.T) {
	runTests(t, []Test{
		That("resolve for").Puts("special"),
		That("resolve put").Puts("$put~"),
		That("fn f { }; resolve f").Puts("$f~"),
		That("use lorem; resolve lorem:put-name").Puts(
			"$lorem:put-name~"),
		That("resolve cat").Puts("(external cat)"),

		That(`f = (mktemp elvXXXXXX); echo 'put x' > $f
		         -source $f; rm $f`).Puts("x"),
		That(`f = (mktemp elvXXXXXX); echo 'put $x' > $f
		         fn p [x]{ -source $f }; p x; rm $f`).Puts("x"),
	})
}

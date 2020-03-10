package str

import (
	"testing"

	"github.com/elves/elvish/pkg/eval"
)

var That = eval.That

func TestStr(t *testing.T) {
	setup := func(ev *eval.Evaler) { ev.Builtin.AddNs("str", Ns) }
	eval.TestWithSetup(t, setup,
		That(`str:compare abc`).ThrowsAny(),
		That(`str:compare abc abc`).Puts("0"),
		That(`str:compare abc def`).Puts("-1"),
		That(`str:compare def abc`).Puts("1"),

		That(`str:contains abc`).ThrowsAny(),
		That(`str:contains abcd x`).Puts(false),
		That(`str:contains abcd bc`).Puts(true),
		That(`str:contains abcd cde`).Puts(false),

		That(`str:contains-any abc`).ThrowsAny(),
		That(`str:contains-any abcd x`).Puts(false),
		That(`str:contains-any abcd xcy`).Puts(true),

		That(`str:equal-fold abc`).ThrowsAny(),
		That(`str:equal-fold ABC abc`).Puts(true),
		That(`str:equal-fold abc ABC`).Puts(true),
		That(`str:equal-fold abc A`).Puts(false),

		That(`str:has-prefix abc`).ThrowsAny(),
		That(`str:has-prefix abcd ab`).Puts(true),
		That(`str:has-prefix abcd cd`).Puts(false),

		That(`str:has-suffix abc`).ThrowsAny(),
		That(`str:has-suffix abcd ab`).Puts(false),
		That(`str:has-suffix abcd cd`).Puts(true),

		That(`str:index abc`).ThrowsAny(),
		That(`str:index abcd cd`).Puts("2"),
		That(`str:index abcd de`).Puts("-1"),

		That(`str:index-any abc`).ThrowsAny(),
		That(`str:index-any "chicken" "aeiouy"`).Puts("2"),
		That(`str:index-any l33t aeiouy`).Puts("-1"),

		That(`str:last-index abc`).ThrowsAny(),
		That(`str:last-index "elven speak elvish" "elv"`).Puts("12"),
		That(`str:last-index "elven speak elvish" "romulan"`).Puts("-1"),

		That(`str:title abc`).Puts("Abc"),
		That(`str:title "abc def"`).Puts("Abc Def"),
		That(`str:to-lower abc def`).ThrowsAny(),

		That(`str:to-lower abc`).Puts("abc"),
		That(`str:to-lower ABC`).Puts("abc"),
		That(`str:to-lower ABC def`).ThrowsAny(),

		That(`str:to-title "her royal highness"`).Puts("HER ROYAL HIGHNESS"),
		That(`str:to-title "хлеб"`).Puts("ХЛЕБ"),

		That(`str:to-upper abc`).Puts("ABC"),
		That(`str:to-upper ABC`).Puts("ABC"),
		That(`str:to-upper ABC def`).ThrowsAny(),

		That(`str:trim "¡¡¡Hello, Elven!!!" "!¡"`).Puts("Hello, Elven"),
		That(`str:trim def`).ThrowsAny(),

		That(`str:trim-left "¡¡¡Hello, Elven!!!" "!¡"`).Puts("Hello, Elven!!!"),
		That(`str:trim-left def`).ThrowsAny(),

		That(`str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hello, "`).Puts("Elven!!!"),
		That(`str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hola, "`).Puts("¡¡¡Hello, Elven!!!"),
		That(`str:trim-prefix def`).ThrowsAny(),

		That(`str:trim-right "¡¡¡Hello, Elven!!!" "!¡"`).Puts("¡¡¡Hello, Elven"),
		That(`str:trim-right def`).ThrowsAny(),

		That(`str:trim-space " \t\n Hello, Elven \n\t\r\n"`).Puts("Hello, Elven"),
		That(`str:trim-space " \t\n Hello  Elven \n\t\r\n"`).Puts("Hello  Elven"),
		That(`str:trim-space " \t\n Hello  Elven \n\t\r\n" argle`).ThrowsAny(),

		That(`str:trim-suffix "¡¡¡Hello, Elven!!!" ", Elven!!!"`).Puts("¡¡¡Hello"),
		That(`str:trim-suffix "¡¡¡Hello, Elven!!!" ", Klingons!!!"`).Puts("¡¡¡Hello, Elven!!!"),
		That(`str:trim-suffix "¡¡¡Hello, Elven!!!"`).ThrowsAny(),
	)
}

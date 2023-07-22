package str

import (
	"fmt"
	"strconv"
	"testing"
	"unicode"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
)

func TestStr(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddNs("str", Ns))
	}
	TestWithEvalerSetup(t, setup,
		That(`str:compare abc`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`str:compare abc abc`).Puts(0),
		That(`str:compare abc def`).Puts(-1),
		That(`str:compare def abc`).Puts(1),

		That(`str:contains abc`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`str:contains abcd x`).Puts(false),
		That(`str:contains abcd bc`).Puts(true),
		That(`str:contains abcd cde`).Puts(false),

		That(`str:contains-any abc`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`str:contains-any abcd x`).Puts(false),
		That(`str:contains-any abcd xcy`).Puts(true),

		That(`str:equal-fold abc`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`str:equal-fold ABC abc`).Puts(true),
		That(`str:equal-fold abc ABC`).Puts(true),
		That(`str:equal-fold abc A`).Puts(false),

		That(`str:fields "abc ABC"`).Puts("abc", "ABC"),
		That(`str:fields "abc    ABC"`).Puts("abc", "ABC"),
		That(`str:fields "  "`).Puts(),

		That(`str:from-codepoints 0x61`).Puts("a"),
		That(`str:from-codepoints 0x4f60 0x597d`).Puts("你好"),
		That(`str:from-codepoints -0x1`).Throws(errs.OutOfRange{
			What:     "codepoint",
			ValidLow: "0", ValidHigh: strconv.Itoa(unicode.MaxRune),
			Actual: "-0x1"}),

		That(fmt.Sprintf(`str:from-codepoints 0x%x`, unicode.MaxRune+1)).Throws(errs.OutOfRange{
			What:     "codepoint",
			ValidLow: "0", ValidHigh: strconv.Itoa(unicode.MaxRune),
			Actual: hex(unicode.MaxRune + 1)}),

		That(`str:from-codepoints 0xd800`).Throws(errs.BadValue{
			What:   "argument to str:from-codepoints",
			Valid:  "valid Unicode codepoint",
			Actual: "0xd800"}),

		That(`str:from-utf8-bytes 0x61`).Puts("a"),
		That(`str:from-utf8-bytes 0xe4 0xbd 0xa0 0xe5 0xa5 0xbd`).Puts("你好"),
		That(`str:from-utf8-bytes -1`).Throws(errs.OutOfRange{
			What:     "byte",
			ValidLow: "0", ValidHigh: "255", Actual: strconv.Itoa(-1)}),

		That(`str:from-utf8-bytes 256`).Throws(errs.OutOfRange{
			What:     "byte",
			ValidLow: "0", ValidHigh: "255", Actual: strconv.Itoa(256)}),

		That(`str:from-utf8-bytes 0xff 0x3 0xaa`).Throws(errs.BadValue{
			What:   "arguments to str:from-utf8-bytes",
			Valid:  "valid UTF-8 sequence",
			Actual: "[255 3 170]"}),

		That(`str:has-prefix abc`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`str:has-prefix abcd ab`).Puts(true),
		That(`str:has-prefix abcd cd`).Puts(false),

		That(`str:has-suffix abc`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`str:has-suffix abcd ab`).Puts(false),
		That(`str:has-suffix abcd cd`).Puts(true),

		That(`str:index abc`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`str:index abcd cd`).Puts(2),
		That(`str:index abcd de`).Puts(-1),

		That(`str:index-any abc`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`str:index-any "chicken" "aeiouy"`).Puts(2),
		That(`str:index-any l33t aeiouy`).Puts(-1),

		That(`str:join : [/usr /bin /tmp]`).Puts("/usr:/bin:/tmp"),
		That(`str:join : ['' a '']`).Puts(":a:"),
		That(`str:join : [(num 1) 2]`).Throws(
			errs.BadValue{What: "input to str:join", Valid: "string", Actual: "number"}),

		That(`str:last-index abc`).Throws(ErrorWithType(errs.ArityMismatch{})),
		That(`str:last-index "elven speak elvish" "elv"`).Puts(12),
		That(`str:last-index "elven speak elvish" "romulan"`).Puts(-1),

		That(`str:replace : / ":usr:bin:tmp"`).Puts("/usr/bin/tmp"),
		That(`str:replace &max=2 : / :usr:bin:tmp`).Puts("/usr/bin:tmp"),

		That(`str:split : /usr:/bin:/tmp`).Puts("/usr", "/bin", "/tmp"),
		That(`str:split : /usr:/bin:/tmp &max=2`).Puts("/usr", "/bin:/tmp"),
		That(`str:split : a:b >&-`).Throws(eval.ErrPortDoesNotSupportValueOutput),

		That(`str:to-codepoints a`).Puts("0x61"),
		That(`str:to-codepoints 你好`).Puts("0x4f60", "0x597d"),
		That(`str:to-codepoints 你好 | str:from-codepoints (all)`).Puts("你好"),
		That(`str:to-codepoints a >&-`).Throws(eval.ErrPortDoesNotSupportValueOutput),

		That(`str:to-utf8-bytes a`).Puts("0x61"),
		That(`str:to-utf8-bytes 你好`).Puts("0xe4", "0xbd", "0xa0", "0xe5", "0xa5", "0xbd"),
		That(`str:to-utf8-bytes 你好 | str:from-utf8-bytes (all)`).Puts("你好"),
		That(`str:to-utf8-bytes a >&-`).Throws(eval.ErrPortDoesNotSupportValueOutput),

		That(`str:title abc`).Puts("Abc"),
		That(`str:title "abc def"`).Puts("Abc Def"),
		That(`str:to-lower abc def`).Throws(ErrorWithType(errs.ArityMismatch{})),

		That(`str:to-lower abc`).Puts("abc"),
		That(`str:to-lower ABC`).Puts("abc"),
		That(`str:to-lower ABC def`).Throws(ErrorWithType(errs.ArityMismatch{})),

		That(`str:to-title "her royal highness"`).Puts("HER ROYAL HIGHNESS"),
		That(`str:to-title "хлеб"`).Puts("ХЛЕБ"),

		That(`str:to-upper abc`).Puts("ABC"),
		That(`str:to-upper ABC`).Puts("ABC"),
		That(`str:to-upper ABC def`).Throws(ErrorWithType(errs.ArityMismatch{})),

		That(`str:trim "¡¡¡Hello, Elven!!!" "!¡"`).Puts("Hello, Elven"),
		That(`str:trim def`).Throws(ErrorWithType(errs.ArityMismatch{})),

		That(`str:trim-left "¡¡¡Hello, Elven!!!" "!¡"`).Puts("Hello, Elven!!!"),
		That(`str:trim-left def`).Throws(ErrorWithType(errs.ArityMismatch{})),

		That(`str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hello, "`).Puts("Elven!!!"),
		That(`str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hola, "`).Puts("¡¡¡Hello, Elven!!!"),
		That(`str:trim-prefix def`).Throws(ErrorWithType(errs.ArityMismatch{})),

		That(`str:trim-right "¡¡¡Hello, Elven!!!" "!¡"`).Puts("¡¡¡Hello, Elven"),
		That(`str:trim-right def`).Throws(ErrorWithType(errs.ArityMismatch{})),

		That(`str:trim-space " \t\n Hello, Elven \n\t\r\n"`).Puts("Hello, Elven"),
		That(`str:trim-space " \t\n Hello  Elven \n\t\r\n"`).Puts("Hello  Elven"),
		That(`str:trim-space " \t\n Hello  Elven \n\t\r\n" argle`).
			Throws(ErrorWithType(errs.ArityMismatch{})),

		That(`str:trim-suffix "¡¡¡Hello, Elven!!!" ", Elven!!!"`).Puts("¡¡¡Hello"),
		That(`str:trim-suffix "¡¡¡Hello, Elven!!!" ", Klingons!!!"`).Puts("¡¡¡Hello, Elven!!!"),
		That(`str:trim-suffix "¡¡¡Hello, Elven!!!"`).Throws(ErrorWithType(errs.ArityMismatch{})),
	)
}

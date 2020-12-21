package eval_test

import (
	"testing"

	. "github.com/elves/elvish/pkg/eval/evaltest"
	"github.com/elves/elvish/pkg/eval/vals"
)

func TestBuiltinFnIO(t *testing.T) {
	Test(t,
		That(`put foo bar`).Puts("foo", "bar"),
		That(`put $nil`).Puts(nil),

		That("print abcd | read-upto c").Puts("abc"),
		// read-upto does not consume more than needed
		That("print abcd | { read-upto c; slurp }").Puts("abc", "d"),
		// read-upto reads up to EOF
		That("print abcd | read-upto z").Puts("abcd"),

		That(`print eof-ending | read-line`).Puts("eof-ending"),
		That(`print "lf-ending\n" | read-line`).Puts("lf-ending"),
		That(`print "crlf-ending\r\n" | read-line`).Puts("crlf-ending"),
		That(`print "extra-cr\r\r\n" | read-line`).Puts("extra-cr\r"),

		That(`print [foo bar]`).Prints("[foo bar]"),
		That(`print foo bar &sep=,`).Prints("foo,bar"),
		That(`echo [foo bar]`).Prints("[foo bar]\n"),
		That(`pprint [foo bar]`).Prints("[\n foo\n bar\n]\n"),
		That(`repr foo bar ['foo bar']`).Prints("foo bar ['foo bar']\n"),

		// A sanity test that show writes something.
		That(`show ?(fail foo) | !=s (slurp) ''`).Puts(true),

		// Baseline for only-{bytes,values}
		That(`{ print bytes; put values }`).Prints("bytes").Puts("values"),
		That(`{ print bytes; put values } | only-bytes`).Prints("bytes").Puts(),
		That(`{ print bytes; put values } | only-values`).Prints("").Puts("values"),

		That(`print "a\nb" | slurp`).Puts("a\nb"),
		That(`print "a\nb" | from-lines`).Puts("a", "b"),
		That(`print "a\nb\n" | from-lines`).Puts("a", "b"),
		That(`echo '{"k": "v", "a": [1, 2]}' '"foo"' | from-json`).
			Puts(vals.MakeMap("k", "v", "a", vals.MakeList(1.0, 2.0)),
				"foo"),
		That(`echo '[null, "foo"]' | from-json`).Puts(
			vals.MakeList(nil, "foo")),
		That(`echo 'invalid' | from-json`).Throws(AnyError),

		That(`put "l\norem" ipsum | to-lines`).Prints("l\norem\nipsum\n"),
		That(`put [&k=v &a=[1 2]] foo | to-json`).
			Prints(`{"a":["1","2"],"k":"v"}
"foo"
`),
		That(`put [$nil foo] | to-json`).Prints("[null,\"foo\"]\n"),
	)
}

func TestBuiltinFnPrintf(t *testing.T) {
	Test(t,
		That(`printf abcd`).Prints("abcd"),
		That(`printf '%s\n%s\n' abc xyz`).Prints("abc\\nxyz\\n"),
		That(`printf "%s\n%s\n" abc xyz`).Prints("abc\nxyz\n"),
		That(`printf '%.1f' 3.1415`).Prints("3.1"),
		That(`printf '%.1f' (float64 3.1415)`).Prints("3.1"),
		That(`printf '%5.3s' 3.1415`).Prints("  3.1"),
		That(`printf '%5.3s' (float64 3.1415)`).Prints("  3.1"),
		That(`printf '%3d' (float64 5)`).Prints("  5"),
		That(`printf '%3d' 5`).Prints("  5"),
		That(`printf '%08b' (float64 5)`).Prints("00000101"),
		That(`printf '%08b' 5`).Prints("00000101"),
		That(`printf '%t' $true`).Prints("true"),

		// Verify that corner cases produce the expected error output.
		That(`printf '%f' 1.3x`).Prints("%!f(cannot parse as number: 1.3x)"),
		That(`printf '%d' 3.5`).Prints("%!d(cannot parse as integer: 3.5)"),
		That(`printf '%d' (float64 5.1)`).Prints("%!d(must be an integer)"),
	)
}

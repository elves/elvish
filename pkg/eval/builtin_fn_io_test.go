package eval

import (
	"testing"

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

		That(`print [foo bar]`).Prints("[foo bar]"),
		That(`print foo bar &sep=,`).Prints("foo,bar"),
		That(`echo [foo bar]`).Prints("[foo bar]\n"),
		That(`pprint [foo bar]`).Prints("[\n foo\n bar\n]\n"),
		That(`repr foo bar ['foo bar']`).Prints("foo bar ['foo bar']\n"),

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
		That(`echo 'invalid' | from-json`).Errors(),

		That(`put "l\norem" ipsum | to-lines`).Prints("l\norem\nipsum\n"),
		That(`put [&k=v &a=[1 2]] foo | to-json`).
			Prints(`{"a":["1","2"],"k":"v"}
"foo"
`),
		That(`put [$nil foo] | to-json`).Prints("[null,\"foo\"]\n"),
	)
}

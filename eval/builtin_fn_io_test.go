package eval

import (
	"testing"

	"github.com/elves/elvish/eval/vals"
)

func TestBuiltinFnIO(t *testing.T) {
	test(t, []TestCase{
		That(`put foo bar`).Puts("foo", "bar"),

		That(`print [foo bar]`).Prints("[foo bar]"),
		That(`print foo bar &sep=,`).Prints("foo,bar"),
		That(`echo [foo bar]`).Prints("[foo bar]\n"),
		That(`pprint [foo bar]`).Prints("[\n foo\n bar\n]\n"),
		That(`repr foo bar ['foo bar']`).Prints("foo bar ['foo bar']\n"),

		That(`print "a\nb" | slurp`).Puts("a\nb"),
		That(`print "a\nb" | from-lines`).Puts("a", "b"),
		That(`print "a\nb\n" | from-lines`).Puts("a", "b"),
		That(`echo '{"k": "v", "a": [1, 2]}' '"foo"' | from-json`).Puts(
			vals.MakeMap(map[interface{}]interface{}{
				"k": "v",
				"a": vals.MakeList("1", "2")}),
			"foo"),
		That(`echo 'invalid' | from-json`).Errors(),

		That(`put "l\norem" ipsum | to-lines`).Prints("l\norem\nipsum\n"),
		That(`put [&k=v &a=[1 2]] foo | to-json`).Prints(
			`{"a":["1","2"],"k":"v"}
"foo"
`),
	})
}

package eval

import "testing"

func TestBuiltinFnIO(t *testing.T) {
	runTests(t, []Test{
		That(`put foo bar`).Puts("foo", "bar"),

		That(`print [foo bar]`).Prints("[foo bar]"),
		That(`print foo bar &sep=,`).Prints("foo,bar"),
		That(`echo [foo bar]`).Prints("[foo bar]\n"),
		That(`pprint [foo bar]`).Prints("[\n foo\n bar\n]\n"),
		That(`repr foo bar ['foo bar']`).Prints("foo bar ['foo bar']\n"),

		That(`print "a\nb" | slurp`).Puts("a\nb"),
		That(`print "a\nb" | from-lines`).Puts("a", "b"),
		That(`print "a\nb\n" | from-lines`).Puts("a", "b"),
		That(`echo 'invalid' | from-json`).Errors(),

		That(`put "l\norem" ipsum | to-lines`).Prints("l\norem\nipsum\n"),
		That(`put [&k=v &a=[1 2]] foo | to-json`).Prints(
			`{"a":["1","2"],"k":"v"}
"foo"
`),
		That(`put [&k=v &a=[1 2.5 $-json-nil $false $true]] foo | to-json &-numberify`).Prints(
			`{"a":[1,2.5,null,false,true],"k":"v"}
"foo"
`),
		That(`echo '{"a":[1,2.5,null,false,true],"k":"v"}' |from-json |to-json &-numberify`).Prints(
			`{"a":[1,2.5,null,false,true],"k":"v"}
`),
	})
}

package eval_test

import (
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

func TestPut(t *testing.T) {
	Test(t,
		That(`put foo bar`).Puts("foo", "bar"),
		That(`put $nil`).Puts(nil),
	)
}

func TestReadUpto(t *testing.T) {
	Test(t,
		That("print abcd | read-upto c").Puts("abc"),
		// read-upto does not consume more than needed
		That("print abcd | { read-upto c; slurp }").Puts("abc", "d"),
		// read-upto reads up to EOF
		That("print abcd | read-upto z").Puts("abcd"),
		That("print abcd | read-upto cd").Throws(eval.ErrInvalidTerminator),
	)
}

func TestReadLine(t *testing.T) {
	Test(t,
		That(`print eof-ending | read-line`).Puts("eof-ending"),
		That(`print "lf-ending\n" | read-line`).Puts("lf-ending"),
		That(`print "crlf-ending\r\n" | read-line`).Puts("crlf-ending"),
		That(`print "extra-cr\r\r\n" | read-line`).Puts("extra-cr\r"),
	)
}

func TestPrint(t *testing.T) {
	Test(t,
		That(`print [foo bar]`).Prints("[foo bar]"),
		That(`print foo bar &sep=,`).Prints("foo,bar"),
	)
}

func TestEcho(t *testing.T) {
	Test(t,
		That(`echo [foo bar]`).Prints("[foo bar]\n"),
	)
}

func TestPprint(t *testing.T) {
	Test(t,
		That(`pprint [foo bar]`).Prints("[\n foo\n bar\n]\n"),
	)
}

func TestReprCmd(t *testing.T) {
	Test(t,
		That(`repr foo bar ['foo bar']`).Prints("foo bar ['foo bar']\n"),
	)
}

func TestShow(t *testing.T) {
	Test(t,
		// A sanity test that show writes something.
		That(`show ?(fail foo) | !=s (slurp) ''`).Puts(true),
	)
}

func TestOnlyBytesAndOnlyValues(t *testing.T) {
	Test(t,
		// Baseline for only-{bytes,values}
		That(`{ print bytes; put values }`).Prints("bytes").Puts("values"),
		That(`{ print bytes; put values } | only-bytes`).Prints("bytes").Puts(),
		That(`{ print bytes; put values } | only-values`).Prints("").Puts("values"),
	)
}

func TestSlurp(t *testing.T) {
	Test(t,
		That(`print "a\nb" | slurp`).Puts("a\nb"),
	)
}

func TestFromLines(t *testing.T) {
	Test(t,
		That(`print "a\nb" | from-lines`).Puts("a", "b"),
		That(`print "a\nb\n" | from-lines`).Puts("a", "b"),
	)
}

func TestFromTerminated(t *testing.T) {
	Test(t,
		That(`print "a\nb\x00\x00c\x00d" | from-terminated "\x00"`).Puts("a\nb", "", "c", "d"),
		That(`print "a\x00b\x00" | from-terminated "\x00"`).Puts("a", "b"),
		That(`print aXbXcXXd | from-terminated "X"`).Puts("a", "b", "c", "", "d"),
		That(`from-terminated "xyz"`).Throws(eval.ErrInvalidTerminator),
	)
}

func TestFromJson(t *testing.T) {
	Test(t,
		That(`echo '{"k": "v", "a": [1, 2]}' '"foo"' | from-json`).
			Puts(vals.MakeMap("k", "v", "a", vals.MakeList(1.0, 2.0)),
				"foo"),
		That(`echo '[null, "foo"]' | from-json`).Puts(
			vals.MakeList(nil, "foo")),
		That(`echo 'invalid' | from-json`).Throws(AnyError),
	)
}

func TestToLines(t *testing.T) {
	Test(t,
		That(`put "l\norem" ipsum | to-lines`).Prints("l\norem\nipsum\n"),
	)
}

func TestToTerminated(t *testing.T) {
	Test(t,
		That(`put "l\norem" ipsum | to-terminated "\x00"`).Prints("l\norem\x00ipsum\x00"),
		That(`to-terminated "X" [a b c]`).Prints("aXbXcX"),
		That(`to-terminated "XYZ" [a b c]`).Throws(eval.ErrInvalidTerminator),
	)
}

func TestToJson(t *testing.T) {
	Test(t,
		That(`put [&k=v &a=[1 2]] foo | to-json`).
			Prints(`{"a":["1","2"],"k":"v"}
"foo"
`),
		That(`put [$nil foo] | to-json`).Prints("[null,\"foo\"]\n"),
	)
}

func TestPrintf(t *testing.T) {
	Test(t,
		That(`printf abcd`).Prints("abcd"),

		That(`printf "%s\n%s\n" abc xyz`).Prints("abc\nxyz\n"),
		That(`printf "%q" "abc xyz"`).Prints(`'abc xyz'`),
		That(`printf "%q" ['a b']`).Prints(`['a b']`),
		That(`printf "%v" abc`).Prints("abc"),
		That(`printf "%#v" "abc xyz"`).Prints(`'abc xyz'`),
		That(`printf '%5.3s' 3.1415`).Prints("  3.1"),
		That(`printf '%5.3s' (float64 3.1415)`).Prints("  3.1"),

		That(`printf '%t' $true`).Prints("true"),
		That(`printf '%t' $nil`).Prints("false"),

		That(`printf '%3d' (num 5)`).Prints("  5"),
		That(`printf '%3d' 5`).Prints("  5"),
		That(`printf '%08b' (num 5)`).Prints("00000101"),
		That(`printf '%08b' 5`).Prints("00000101"),

		That(`printf '%.1f' 3.1415`).Prints("3.1"),
		That(`printf '%.1f' (float64 3.1415)`).Prints("3.1"),

		// Does not interpret escape sequences
		That(`printf '%s\n%s\n' abc xyz`).Prints("abc\\nxyz\\n"),

		// Error cases

		// Float verb with argument that can't be converted to float
		That(`printf '%f' 1.3x`).Prints("%!f(cannot parse as number: 1.3x)"),
		// Integer verb with argument that can't be converted to integer
		That(`printf '%d' 3.5`).Prints("%!d(cannot parse as integer: 3.5)"),
		// Unsupported verb
		That(`printf '%A' foo`).Prints("%!A(unsupported formatting verb)"),
	)
}

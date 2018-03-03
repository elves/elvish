package eval

import (
	"os"
	"testing"

	"github.com/elves/elvish/eval/vals"
)

func TestBuiltinFnMisc(t *testing.T) {
	runTests(t, []Test{
		NewTest("resolve for").WantOutStrings("special"),
		NewTest("resolve put").WantOutStrings("$put~"),
		NewTest("fn f { }; resolve f").WantOutStrings("$f~"),
		NewTest("use lorem; resolve lorem:put-name").WantOutStrings(
			"$lorem:put-name~"),
		NewTest("resolve cat").WantOutStrings("(external cat)"),

		NewTest(`f = (mktemp elvXXXXXX); echo 'put x' > $f
		         -source $f; rm $f`).WantOut("x"),
		NewTest(`f = (mktemp elvXXXXXX); echo 'put $x' > $f
		         fn p [x]{ -source $f }; p x; rm $f`).WantOut("x"),
	})
}

func TestBuiltinFnEnv(t *testing.T) {
	oldpath := os.Getenv("PATH")
	listSep := string(os.PathListSeparator)
	runTests(t, []Test{
		{`getenv var`, want{err: ErrMissingEnvVar}},
		{`setenv var test1`, want{}},
		{`getenv var`, want{out: strs("test1")}},
		{`put $E:var`, want{out: strs("test1")}},
		{`setenv var test2`, want{}},
		{`getenv var`, want{out: strs("test2")}},
		{`put $E:var`, want{out: strs("test2")}},

		{`setenv PATH /test-path`, want{}},
		{`put $paths`, want{out: []interface{}{
			vals.MakeList(strs("/test-path")...)}}},
		{`paths = [/test-path2 $@paths]`, want{}},
		{`getenv PATH`, want{out: strs(
			"/test-path2" + listSep + "/test-path")}},
	})
	os.Setenv("PATH", oldpath)
}

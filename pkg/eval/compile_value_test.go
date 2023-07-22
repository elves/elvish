package eval_test

import (
	"errors"
	"fmt"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/testutil"
)

func TestCompound(t *testing.T) {
	Test(t,
		That("put {fi,elvi}sh{1.0,1.1}").Puts(
			"fish1.0", "fish1.1", "elvish1.0", "elvish1.1"),

		// As a special case, an empty compound expression evaluates to an empty
		// string.
		That("put {}").Puts(""),
		That("put [&k=][k]").Puts(""),

		// TODO: Test the case where fsutil.GetHome returns an error.

		// Error in any of the components throws an exception.
		That("put a{[][1]}").Throws(ErrorWithType(errs.OutOfRange{}), "[][1]"),
		// Error in concatenating the values throws an exception.
		That("put []a").Throws(ErrorWithMessage("cannot concatenate list and string")),
		// Error when applying tilde throws an exception.
		That("put ~[]").Throws(ErrorWithMessage("tilde doesn't work on value of type list")),
	)
}

func TestIndexing(t *testing.T) {
	Test(t,
		That("put [a b c][2]").Puts("c"),
		That("put [][0]").Throws(ErrorWithType(errs.OutOfRange{}), "[][0]"),
		That("put [&key=value][key]").Puts("value"),
		That("put [&key=value][bad]").Throws(
			vals.NoSuchKey("bad"), "[&key=value][bad]"),

		That("put (fail x)[a]").Throws(FailError{"x"}, "fail x"),
		That("put [foo][(fail x)]").Throws(FailError{"x"}, "fail x"),
	)
}

func TestListLiteral(t *testing.T) {
	Test(t,
		That("put [a b c]").Puts(vals.MakeList("a", "b", "c")),
		That("put []").Puts(vals.EmptyList),
		// List expression errors if an element expression errors.
		That("put [ [][0] ]").Throws(ErrorWithType(errs.OutOfRange{}), "[][0]"),
	)
}

func TestMapLiteral(t *testing.T) {
	Test(t,
		That("put [&key=value]").Puts(vals.MakeMap("key", "value")),
		That("put [&]").Puts(vals.EmptyMap),
		// Map keys and values may evaluate to multiple values as long as their
		// numbers match.
		That("put [&{a b}={foo bar}]").Puts(vals.MakeMap("a", "foo", "b", "bar")),
		// Map expression errors if a key or value expression errors.
		That("put [ &[][0]=a ]").Throws(ErrorWithType(errs.OutOfRange{}), "[][0]"),
		That("put [ &a=[][0] ]").Throws(ErrorWithType(errs.OutOfRange{}), "[][0]"),
		// Map expression errors if number of keys and values in a single pair
		// does not match.
		That("put [&{a b}={foo bar lorem}]").Throws(ErrorWithMessage("2 keys but 3 values")),
	)
}

func TestStringLiteral(t *testing.T) {
	Test(t,
		That(`put 'such \"''literal'`).Puts(`such \"'literal`),
		That(`put "much \n\033[31;1m$cool\033[m"`).Puts("much \n\033[31;1m$cool\033[m"),
	)
}

func TestTilde(t *testing.T) {
	home := testutil.InTempHome(t)
	testutil.ApplyDir(testutil.Dir{"file1": "", "file2": ""})

	otherHome := testutil.TempDir(t)
	testutil.ApplyDirIn(testutil.Dir{"other1": "", "other2": ""}, otherHome)
	testutil.Set(t, GetHome, func(name string) (string, error) {
		switch name {
		case "":
			return fsutil.GetHome("")
		case "other":
			return otherHome, nil
		default:
			return "", fmt.Errorf("don't know home of %v", name)
		}
	})

	Test(t,
		That("put ~").Puts(home),
		That("put ~/src").Puts(home+"/src"),
		// Make sure that tilde processing retains trailing slashes.
		That("put ~/src/").Puts(home+"/src/"),
		// Tilde and wildcard.
		That("put ~/*").Puts(home+"/file1", home+"/file2"),

		// Regression test for b.elv.sh/1246.
		That("put ~other").Puts(otherHome),
		// Regression test for #793.
		That("put ~other/*").Puts(otherHome+"/other1", otherHome+"/other2"),

		That("put ~bad/*").Throws(ErrorWithMessage("don't know home of bad"), "~bad/*"),

		// TODO: This should be a compilation error.
		That("put ~*").Throws(ErrCannotDetermineUsername, "~*"),
	)
}

func TestTilde_ErrorForCurrentUser(t *testing.T) {
	err := errors.New("fake error")
	testutil.Set(t, GetHome, func(name string) (string, error) { return "", err })

	Test(t,
		That("put ~").Throws(err, "~"),
		That("put ~/foo").Throws(err, "~/foo"),
		That("put ~/*").Throws(err, "~/*"),
	)
}

func TestWildcard(t *testing.T) {
	Test(t,
		That("put ***").DoesNotCompile(`bad wildcard: "***"`),
	)
	// More tests in glob_test.go
}

func TestOutputCapture(t *testing.T) {
	Test(t,
		// Output capture
		That("put (put lorem ipsum)").Puts("lorem", "ipsum"),
		That("put (print \"lorem\nipsum\")").Puts("lorem", "ipsum"),
		// \r\n is also supported as a line separator
		That(`print "lorem\r\nipsum\r\n" | all`).Puts("lorem", "ipsum"),
	)
}

func TestExceptionCapture(t *testing.T) {
	Test(t,
		// Exception capture
		That("bool ?(nop); bool ?(e:false)").Puts(true, false),
	)
}

func TestVariableUse(t *testing.T) {
	Test(t,
		That("var x = foo", "put $x").Puts("foo"),
		// Must exist before use
		That("put $x").DoesNotCompile("variable $x not found"),
		That("put $x[0]").DoesNotCompile("variable $x not found"),
		// Compounding
		That("var x = SHELL", "put 'WOW, SUCH '$x', MUCH COOL'\n").
			Puts("WOW, SUCH SHELL, MUCH COOL"),
		// Splicing
		That("var x = [elvish rules]", "put $@x").Puts("elvish", "rules"),

		// Variable namespace
		// ------------------

		// Unqualified name resolves to local name before upvalue.
		That("var x = outer; { var x = inner; put $x }").Puts("inner"),
		// Unqualified name resolves to upvalue if no local name exists.
		That("var x = outer; { put $x }").Puts("outer"),
		// Unqualified name resolves to builtin if no local name or upvalue
		// exists.
		That("put $true").Puts(true),
		// Names like $:foo are reserved for now.
		That("var x = val; put $:x").DoesNotCompile("variable $:x not found"),

		// Pseudo-namespace E: provides read-write access to environment
		// variables. Colons inside the name are supported.
		That("set-env a:b VAL; put $E:a:b").Puts("VAL"),
		That("set E:a:b = VAL2; get-env a:b").Puts("VAL2"),

		// Pseudo-namespace e: provides readonly access to external commands.
		// Only names ending in ~ are resolved, and resolution always succeeds
		// regardless of whether the command actually exists. Colons inside the
		// name are supported.
		That("put $e:a:b~").Puts(NewExternalCmd("a:b")),

		// A "normal" namespace access indexes the namespace as a variable.
		That("var ns: = (ns [&a= val]); put $ns:a").Puts("val"),
		// Multi-level namespace access is supported.
		That("var ns: = (ns [&a:= (ns [&b= val])]); put $ns:a:b").Puts("val"),
	)
}

func TestVariableUse_NonExistentVariableInModule(t *testing.T) {
	TestWithEvalerSetup(t, func(ev *Evaler) { ev.AddModule("mod", &Ns{}) },
		// Variables in other modules are checked at runtime.
		That("use mod; put $mod:bad").Throws(
			ErrorWithMessage("variable $mod:bad not found"), "$mod:bad"),
	)
}

func TestClosure(t *testing.T) {
	Test(t,
		That("{|| }").DoesNothing(),
		That("{|x| put $x} foo").Puts("foo"),

		// Assigning to captured variable
		That("var x = lorem; {|| set x = ipsum}; put $x").Puts("ipsum"),
		That("var x = lorem; {|| put $x; set x = ipsum }; put $x").
			Puts("lorem", "ipsum"),

		// Assigning to element of captured variable
		That("var x = a; { set x = b }; put $x").Puts("b"),
		That("var x = [a]; { set x[0] = b }; put $x[0]").Puts("b"),

		// Shadowing
		That("var x = ipsum; { var x = lorem; put $x }; put $x").
			Puts("lorem", "ipsum"),

		// Shadowing by argument
		That("var x = ipsum; {|x| put $x; set x = BAD } lorem; put $x").
			Puts("lorem", "ipsum"),

		// Closure captures new local variables every time
		That("fn f { var x = (num 0); put { set x = (+ $x 1) } { put $x } }",
			"var inc1 put1 = (f); $put1; $inc1; $put1",
			"var inc2 put2 = (f); $put2; $inc2; $put2").Puts(0, 1, 0, 1),

		// Rest argument.
		That("{|x @xs| put $x $xs } a b c").Puts("a", vals.MakeList("b", "c")),
		That("{|a @b c| put $a $b $c } a b c d").
			Puts("a", vals.MakeList("b", "c"), "d"),
		// Options.
		That("{|a &k=v| put $a $k } foo &k=bar").Puts("foo", "bar"),
		// Option default value.
		That("{|a &k=v| put $a $k } foo").Puts("foo", "v"),
		// Option must have default value
		That("{|&k| }").DoesNotCompile("option must have default value"),
		// Exception when evaluating option default value.
		That("{|&a=[][0]| }").Throws(ErrorWithType(errs.OutOfRange{}), "[][0]"),
		// Option default value must be one value.
		That("{|&a=(put foo bar)| }").Throws(
			errs.ArityMismatch{What: "option default value", ValidLow: 1, ValidHigh: 1, Actual: 2},
			"(put foo bar)"),

		// Argument name must be unqualified.
		That("{|a:b| }").DoesNotCompile("argument name must be unqualified"),
		// Argument name must not be empty.
		That("{|''| }").DoesNotCompile("argument name must not be empty"),
		That("{|@| }").DoesNotCompile("argument name must not be empty"),
		// Option name must be unqualified.
		That("{|&a:b=1| }").DoesNotCompile("option name must be unqualified"),
		// Option name must not be empty.
		That("{|&''=b| }").DoesNotCompile("option name must not be empty"),
		// Should not have multiple rest arguments.
		That("{|@a @b| }").DoesNotCompile("only one argument may have @ prefix"),
	)
}

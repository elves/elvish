package eval

import (
	"testing"

	"github.com/elves/elvish/eval/vals"
)

var valueTests = []Test{
	// Compounding
	// -----------
	That("put {fi,elvi}sh{1.0,1.1}").Puts(
		"fish1.0", "fish1.1", "elvish1.0", "elvish1.1"),

	// List, Map and Indexing
	// ----------------------

	That("echo [a b c] [&key=value] | each $put~").Puts(
		"[a b c] [&key=value]"),
	That("put [a b c][2]").Puts("c"),
	That("put [&key=value][key]").Puts("value"),

	// String Literals
	// ---------------
	That(`put 'such \"''literal'`).Puts(`such \"'literal`),
	That(`put "much \n\033[31;1m$cool\033[m"`).Puts(
		"much \n\033[31;1m$cool\033[m"),

	// Captures
	// ---------

	// Output capture
	That("put (put lorem ipsum)").Puts("lorem", "ipsum"),
	That("put (print \"lorem\nipsum\")").Puts("lorem", "ipsum"),

	// Exception capture
	That("bool ?(nop); bool ?(e:false)").Puts(true, false),

	// Variable Use
	// ------------

	// Compounding
	That("x='SHELL'\nput 'WOW, SUCH '$x', MUCH COOL'\n").Puts(
		"WOW, SUCH SHELL, MUCH COOL"),
	// Splicing
	That("x=[elvish rules]; put $@x").Puts("elvish", "rules"),

	// Wildcard; see testmain_test.go for FS setup
	// -------------------------------------------

	That("put *").PutsStrings(fileListing),
	That("put a/b/nonexistent*").ErrorsWith(ErrWildcardNoMatch),
	That("put a/b/nonexistent*[nomatch-ok]").DoesNothing(),

	// Character set and range
	That("put ?[set:ab]*").PutsStrings(getFilesWithPrefix("a", "b")),
	That("put ?[range:a-c]*").PutsStrings(getFilesWithPrefix("a", "b", "c")),
	That("put ?[range:a~c]*").PutsStrings(getFilesWithPrefix("a", "b")),
	That("put *[range:a-z]").Puts("bar", "dir", "foo", "ipsum", "lorem"),

	// Exclusion
	That("put *[but:foo][but:lorem]").PutsStrings(getFilesBut("foo", "lorem")),

	// Tilde
	// -----
	That("h=$E:HOME; E:HOME=/foo; put ~ ~/src; E:HOME=$h").Puts("/foo", "/foo/src"),

	// Closure
	// -------

	That("[]{ }").DoesNothing(),
	That("[x]{put $x} foo").Puts("foo"),

	// Variable capture
	That("x=lorem; []{x=ipsum}; put $x").Puts("ipsum"),
	That("x=lorem; []{ put $x; x=ipsum }; put $x").Puts("lorem", "ipsum"),

	// Shadowing
	That("x=ipsum; []{ local:x=lorem; put $x }; put $x").Puts("lorem", "ipsum"),

	// Shadowing by argument
	That("x=ipsum; [x]{ put $x; x=BAD } lorem; put $x").Puts("lorem", "ipsum"),

	// Closure captures new local variables every time
	That(`fn f []{ x=0; put []{x=(+ $x 1)} []{put $x} }
		      {inc1,put1}=(f); $put1; $inc1; $put1
			  {inc2,put2}=(f); $put2; $inc2; $put2`).Puts("0", "1", "0", "1"),

	// Rest argument.
	That("[x @xs]{ put $x $xs } a b c").Puts("a", vals.MakeList("b", "c")),
	// Options.
	That("[a &k=v]{ put $a $k } foo &k=bar").Puts("foo", "bar"),
	// Option default value.
	That("[a &k=v]{ put $a $k } foo").Puts("foo", "v"),
}

func TestValue(t *testing.T) {
	runTests(t, valueTests)
}

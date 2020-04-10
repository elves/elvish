package eval

import (
	"sort"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/util"
)

func TestCompileValue(t *testing.T) {
	home, cleanup := InTempHome()
	mustCreateEmpty("file1")
	mustCreateEmpty("file2")
	defer cleanup()

	Test(t,
		// Compounding
		// -----------
		That("put {fi,elvi}sh{1.0,1.1}").Puts(
			"fish1.0", "fish1.1", "elvish1.0", "elvish1.1"),

		// As a special case, an empty compound expression evaluates to an empty
		// string.
		That("put {}").Puts(""),
		That("put [&k=][k]").Puts(""),

		// TODO: Test the case where util.GetHome returns an error.

		// Error in any of the components throws an exception.
		That("put a{[][1]}").ThrowsMessage("index out of range"),
		// Error in concatenating the values throws an exception.
		That("put []a").ThrowsMessage("cannot concatenate list and string"),
		// Error when applying tilde throws an exception.
		That("put ~[]").ThrowsMessage("tilde doesn't work on value of type list"),

		// List, Map and Indexing
		// ----------------------

		That("echo [a b c] [&key=value] | each $put~").
			Puts("[a b c] [&key=value]"),
		That("put [a b c][2]").Puts("c"),
		That("put [&key=value][key]").Puts("value"),

		// Map keys and values may evaluate to multiple values as long as their
		// numbers match.
		That("put [&{a b}={foo bar}]").Puts(vals.MakeMap("a", "foo", "b", "bar")),

		// List expression errors if an element expression errors.
		That("put [ [][0] ]").ThrowsMessage("index out of range"),
		// Map expression errors if a key or value expression errors.
		That("put [ &[][0]=a ]").ThrowsMessage("index out of range"),
		That("put [ &a=[][0] ]").ThrowsMessage("index out of range"),
		// Map expression errors if number of keys and values in a single pair
		// does not match.
		That("put [&{a b}={foo bar lorem}]").ThrowsMessage("2 keys but 3 values"),

		// String Literals
		// ---------------
		That(`put 'such \"''literal'`).Puts(`such \"'literal`),
		That(`put "much \n\033[31;1m$cool\033[m"`).
			Puts("much \n\033[31;1m$cool\033[m"),

		// Captures
		// ---------

		// Output capture
		That("put (put lorem ipsum)").Puts("lorem", "ipsum"),
		That("put (print \"lorem\nipsum\")").Puts("lorem", "ipsum"),

		// Exception capture
		That("bool ?(nop); bool ?(e:false)").Puts(true, false),

		// Variable Use
		// ------------

		That("x = foo", "put $x").Puts("foo"),
		// Must exist before use
		That("put $x").DoesNotCompile(),
		That("put $x[0]").DoesNotCompile(),
		// Compounding
		That("x = SHELL", "put 'WOW, SUCH '$x', MUCH COOL'\n").
			Puts("WOW, SUCH SHELL, MUCH COOL"),
		// Splicing
		That("x = [elvish rules]", "put $@x").Puts("elvish", "rules"),

		// Variable namespace
		// ------------------

		// Pseudo-namespace local: accesses the local scope.
		That("x = outer; { local:x = inner; put $local:x }").Puts("inner"),
		// Pseudo-namespace up: accesses upvalues.
		That("x = outer; { local:x = inner; put $up:x }").Puts("outer"),
		// Pseudo-namespace builtin: accesses builtins.
		That("put $builtin:true").Puts(true),
		// Unqualified name prefers local: to up:.
		That("x = outer; { local:x = inner; put $x }").Puts("inner"),
		// Unqualified name resolves to upvalue if no local name exists.
		That("x = outer; { put $x }").Puts("outer"),
		// Unqualified name resolves to builtin if no local name or upvalue
		// exists.
		That("put $true").Puts(true),
		// A name can be explicitly unqualified by having a leading colon.
		That("x = val; put $:x").Puts("val"),
		That("put $:true").Puts(true),

		// Pseudo-namespace E: provides read-write access to environment
		// variables. Colons inside the name are supported.
		That("set-env a:b VAL; put $E:a:b").Puts("VAL"),
		That("E:a:b = VAL2; get-env a:b").Puts("VAL2"),

		// Pseudo-namespace e: provides readonly access to external commands.
		// Only names ending in ~ are resolved, and resolution always succeeds
		// regardless of whether the command actually exists. Colons inside the
		// name are supported.
		That("put $e:a:b~").Puts(ExternalCmd{Name: "a:b"}),

		// A "normal" namespace access indexes the namespace as a variable.
		That("ns: = (ns [&a= val]); put $ns:a").Puts("val"),
		// Multi-level namespace access is supported.
		That("ns: = (ns [&a:= (ns [&b= val])]); put $ns:a:b").Puts("val"),
		// Multi-level namespace access can have a leading colon to signal that
		// the first component is unqualified.
		That("ns: = (ns [&a:= (ns [&b= val])]); put $:ns:a:b").Puts("val"),
		// Multi-level namespace access can be combined with the local:
		// pseudo-namespaces.
		That("ns: = (ns [&a:= (ns [&b= val])]); put $local:ns:a:b").Puts("val"),
		// Multi-level namespace access can be combined with the up:
		// pseudo-namespaces.
		That("ns: = (ns [&a:= (ns [&b= val])]); { put $up:ns:a:b }").Puts("val"),

		// Tilde
		// -----
		That("put ~").Puts(home),
		That("put ~/src").Puts(home+"/src"),
		// Make sure that tilde processing retains trailing slashes.
		That("put ~/src/").Puts(home+"/src/"),
		// Tilde and wildcard.
		That("put ~/*").Puts(home+"/file1", home+"/file2"),
		// TODO(xiaq): Add regression test for #793.

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
		  {inc2,put2}=(f); $put2; $inc2; $put2`).Puts("0", 1.0, "0", 1.0),

		// Rest argument.
		That("[x @xs]{ put $x $xs } a b c").Puts("a", vals.MakeList("b", "c")),
		// Options.
		That("[a &k=v]{ put $a $k } foo &k=bar").Puts("foo", "bar"),
		// Option default value.
		That("[a &k=v]{ put $a $k } foo").Puts("foo", "v"),

		// Argument name must be unqualified.
		That("[a:b]{ }").DoesNotCompile(),
		// Argument name must not be empty.
		That("['']{ }").DoesNotCompile(),
		That("[@]{ }").DoesNotCompile(),
		// Only the last argument may be prefixed with @.
		That("[@a b]{ }").DoesNotCompile(),
		// Option name must be unqualified.
		That("[&a:b=1]{ }").DoesNotCompile(),
		// Option name must not be empty.
		That("[&''=b]{ }").DoesNotCompile(),

		// Option default value must be one value.
		That("[&a=(put foo bar)]{ }").ThrowsMessage("option default value must be a single value; got 2 values"),
	)
}

var (
	filesToCreate = []string{
		"a1", "a2", "a3", "a10",
		"b1", "b2", "b3",
		"c1", "c2",
		"foo", "bar", "lorem", "ipsum",
	}
	dirsToCreate = []string{"dir", "dir2"}
	fileListing  = getFileListing()
)

func getFileListing() []string {
	var x []string
	x = append(x, filesToCreate...)
	x = append(x, dirsToCreate...)
	sort.Strings(x)
	return x
}

func getFilesWithPrefix(prefixes ...string) []string {
	var x []string
	for _, name := range fileListing {
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				x = append(x, name)
				break
			}
		}
	}
	sort.Strings(x)
	return x
}

func getFilesBut(excludes ...string) []string {
	var x []string
	for _, name := range fileListing {
		excluded := false
		for _, exclude := range excludes {
			if name == exclude {
				excluded = true
				break
			}
		}
		if !excluded {
			x = append(x, name)
		}
	}
	sort.Strings(x)
	return x
}

func TestWildcard(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	for _, filename := range filesToCreate {
		mustCreateEmpty(filename)
	}
	for _, dirname := range dirsToCreate {
		mustMkdirAll(dirname, 0700)
	}

	Test(t,
		That("put *").PutsStrings(fileListing),
		That("put a/b/nonexistent*").Throws(ErrWildcardNoMatch),
		That("put a/b/nonexistent*[nomatch-ok]").DoesNothing(),

		// Character set and range
		That("put ?[set:ab]*").PutsStrings(getFilesWithPrefix("a", "b")),
		That("put ?[range:a-c]*").PutsStrings(getFilesWithPrefix("a", "b", "c")),
		That("put ?[range:a~c]*").PutsStrings(getFilesWithPrefix("a", "b")),
		That("put *[range:a-z]").Puts("bar", "dir", "foo", "ipsum", "lorem"),

		// Exclusion
		That("put *[but:foo][but:lorem]").PutsStrings(getFilesBut("foo", "lorem")),
	)
}

package eval

import "testing"

var valueTests = []Test{
	// Compounding
	// -----------
	{"put {fi,elvi}sh{1.0,1.1}",
		want{out: strs("fish1.0", "fish1.1", "elvish1.0", "elvish1.1")}},

	// List, Map and Indexing
	// ----------------------

	{"echo [a b c] [&key=value] | each put",
		want{out: strs("[a b c] [&key=value]")}},
	{"put [a b c][2]", want{out: strs("c")}},
	{"put [&key=value][key]", want{out: strs("value")}},

	// String Literals
	// ---------------
	{`put 'such \"''literal'`, want{out: strs(`such \"'literal`)}},
	{`put "much \n\033[31;1m$cool\033[m"`,
		want{out: strs("much \n\033[31;1m$cool\033[m")}},

	// Captures
	// ---------

	// Output capture
	{"put (put lorem ipsum)", want{out: strs("lorem", "ipsum")}},
	{"put (print \"lorem\nipsum\")", want{out: strs("lorem", "ipsum")}},

	// Exception capture
	{"bool ?(nop); bool ?(e:false)", want{out: bools(true, false)}},

	// Variable Use
	// ------------

	// Compounding
	{"x='SHELL'\nput 'WOW, SUCH '$x', MUCH COOL'\n",
		want{out: strs("WOW, SUCH SHELL, MUCH COOL")}},
	// Splicing
	{"x=[elvish rules]; put $@x", want{out: strs("elvish", "rules")}},

	// Wildcard; see testmain_test.go for FS setup
	// -------------------------------------------

	{"put *", want{out: strs(fileListing...)}},
	{"put a/b/nonexistent*", want{err: ErrWildcardNoMatch}},
	{"put a/b/nonexistent*[nomatch-ok]", wantNothing},

	// Character set and range
	{"put ?[set:ab]*", want{out: strs(getFilesWithPrefix("a", "b")...)}},
	{"put ?[range:a-c]*", want{out: strs(getFilesWithPrefix("a", "b", "c")...)}},
	{"put ?[range:a~c]*", want{out: strs(getFilesWithPrefix("a", "b")...)}},
	{"put *[range:a-z]", want{out: strs("bar", "dir", "foo", "ipsum", "lorem")}},

	// Exclusion
	{"put *[but:foo but:lorem]", want{out: strs(getFilesBut("foo", "lorem")...)}},

	// Tilde
	// -----
	{"h=$E:HOME; E:HOME=/foo; put ~ ~/src; E:HOME=$h",
		want{out: strs("/foo", "/foo/src")}},

	// Closure
	// -------

	{"[]{ }", wantNothing},
	{"[x]{put $x} foo", want{out: strs("foo")}},

	// Variable capture
	{"x=lorem; []{x=ipsum}; put $x", want{out: strs("ipsum")}},
	{"x=lorem; []{ put $x; x=ipsum }; put $x",
		want{out: strs("lorem", "ipsum")}},

	// Shadowing
	{"x=ipsum; []{ local:x=lorem; put $x }; put $x",
		want{out: strs("lorem", "ipsum")}},

	// Shadowing by argument
	{"x=ipsum; [x]{ put $x; x=BAD } lorem; put $x",
		want{out: strs("lorem", "ipsum")}},

	// Closure captures new local variables every time
	{`fn f []{ x=0; put []{x=(+ $x 1)} []{put $x} }
		      {inc1,put1}=(f); $put1; $inc1; $put1
			  {inc2,put2}=(f); $put2; $inc2; $put2`,
		want{out: strs("0", "1", "0", "1")}},

	// Rest argument.
	{"[x @xs]{ put $x $xs } a b c",
		want{out: []Value{String("a"), NewList(String("b"), String("c"))}}},
	// Options.
	{"[a &k=v]{ put $a $k } foo &k=bar", want{out: strs("foo", "bar")}},
	// Option default value.
	{"[a &k=v]{ put $a $k } foo", want{out: strs("foo", "v")}},
}

func TestValue(t *testing.T) {
	RunTests(t, libDir, valueTests)
}

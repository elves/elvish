package eval

import "testing"

var builtinSpecialTests = []Test{
	// del
	NewTest("x = [&k=v &k2=v2]; del x[k2]; keys $x").WantOutStrings("k"),
	NewTest("x = [[&k=v &k2=v2]]; del x[0][k2]; keys $x[0]").WantOutStrings("k"),

	// if
	{"if true { put then }", want{out: strs("then")}},
	{"if $false { put then } else { put else }", want{out: strs("else")}},
	{"if $false { put 1 } elif $false { put 2 } else { put 3 }",
		want{out: strs("3")}},
	{"if $false { put 2 } elif true { put 2 } else { put 3 }",
		want{out: strs("2")}},

	// try
	{"try { nop } except { put bad } else { put good }", want{out: strs("good")}},
	{"try { e:false } except - { put bad } else { put good }", want{out: strs("bad")}},
	// "finally" should not catch the exception
	NewTest("try { fail bad } finally { put final }").WantOut("final").WantAnyErr(),

	// while
	{"x=0; while (< $x 4) { put $x; x=(+ $x 1) }",
		want{out: strs("0", "1", "2", "3")}},
	NewTest("x = 0; while (< $x 4) { put $x; break }").WantOutStrings("0"),
	NewTest("x = 0; while (< $x 4) { fail haha }").WantAnyErr(),

	// for
	{"for x [tempora mores] { put 'O '$x }",
		want{out: strs("O tempora", "O mores")}},
	// break
	{"for x [a] { break } else { put $x }", wantNothing},
	// else
	{"for x [a] { put $x } else { put $x }", want{out: strs("a")}},
	// continue
	{"for x [a b] { put $x; continue; put $x; }", want{out: strs("a", "b")}},

	// fn.
	{"fn f [x]{ put x=$x'.' }; f lorem; f ipsum",
		want{out: strs("x=lorem.", "x=ipsum.")}},
	// return.
	{"fn f []{ put a; return; put b }; f", want{out: strs("a")}},

	// Modules (see testmain_test.go for setup)
	// "use" imports a module.
	{`use lorem; put $lorem:name`, want{out: strs("lorem")}},
	// imports are lexically scoped
	// TODO: Support testing for compilation error
	// {`{ use lorem }; put $lorem:name`, want{err: errAny}},

	// use of imported variable is captured in upvalue
	{`({ use lorem; put { { put $lorem:name } } })`, want{out: strs("lorem")}},
	// use of imported function is also captured in upvalue
	{`{ use lorem; { lorem:put-name } }`, want{out: strs("lorem")}},

	// use of a nested module
	{`use a:b/c/d; put $d:name`, want{out: strs("a/b/c/d")}},
	// module is cached after first use
	{`use has/init; use has/init`, want{out: strs("has/init")}},
	// overriding module
	{`use d; put $d:name; use a/b/c/d; put $d:name`,
		want{out: strs("d", "a/b/c/d")}},
	// relative uses
	{`use a/b/c/x; put $x:d $x:lorem`, want{out: strs("a/b/c/d", "lorem")}},

	// Variables defined in the default global scope is invisible from modules
	NewTest("x = foo; use put-x").WantAnyErr(),

	// TODO: Test module namespace
}

func TestBuiltinSpecial(t *testing.T) {
	runTests(t, builtinSpecialTests)
}

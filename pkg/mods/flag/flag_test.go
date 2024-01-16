package flag

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

func TestCall(t *testing.T) {
	TestWithEvalerSetup(t, setup,
		That("flag:call {|&bool=$false| put $bool } [-bool]").Puts(true),
		That("flag:call {|&str=''| put $str } [-str foo]").Puts("foo"),
		That("flag:call {|&opt=$false arg| put $opt $arg } [-opt foo]").
			Puts(true, "foo"),
		// Flag parsing error
		That("flag:call { } [-bad '']").
			Throws(ErrorWithMessage("flag provided but not defined: -bad")),
		// Bad argument list
		That("flag:call { } [(num 0)]").
			Throws(ErrorWithMessage("wrong type: need string, got number")),
		// Validate invalid default flag value raises an exception.
		That("flag:call {|&f=$nil| } [-f 1]").
			Throws(errs.BadValue{What: "flag default value",
				Valid: "boolean, number, string or list", Actual: "$nil"}),
		// More flag parsing logic is covered in TestParse
	)

}

func TestParse(t *testing.T) {
	TestWithEvalerSetup(t, setup,
		// Different types of flags
		That("flag:parse [-bool] [[bool $false bool]]").
			Puts(vals.MakeMap("bool", true), vals.EmptyList),
		That("flag:parse [-str lorem] [[str '' string]]").
			Puts(vals.MakeMap("str", "lorem"), vals.EmptyList),
		That("flag:parse [-num 100] [[num (num 0) number]]").
			Puts(vals.MakeMap("num", 100), vals.EmptyList),
		That("flag:parse [-list a,b] [[list [] list]]").
			Puts(vals.MakeMap("list", vals.MakeList("a", "b")), vals.EmptyList),
		// Multiple flags, and non-flag arguments
		That("flag:parse [-v -n foo bar] [[v $false verbose] [n '' name]]").
			Puts(vals.MakeMap("v", true, "n", "foo"), vals.MakeList("bar")),

		// Flag parsing error
		That("flag:parse [-bad ''] []").
			Throws(ErrorWithMessage("flag provided but not defined: -bad")),
		// Unsupported type for default value
		That("flag:parse [-map ''] [[map [&] map]]").
			Throws(errs.BadValue{What: "flag default value",
				Valid: "boolean, number, string or list", Actual: "[&]"}),
		// TODO: Improve these errors to point out where the wrong type occurs
		// Bad argument list
		That("flag:parse [(num 0)] []").
			Throws(ErrorWithMessage("wrong type: need string, got number")),
		// Bad spec list
		That("flag:parse [] [(num 0)]").
			Throws(ErrorWithMessage("wrong type: need !!vector.Vector, got number")),
	)
}

func TestParseGetopt(t *testing.T) {
	vFlag := vals.MakeMap(
		"spec", vals.MakeMap("short", "v"), "long", false, "arg", "")

	TestWithEvalerSetup(t, setup,
		// Basic test
		That("flag:parse-getopt [-v foo] [[&short=v]]").
			Puts(vals.MakeList(vFlag), vals.MakeList("foo")),
		// Extra info in spec
		That("flag:parse-getopt [-v foo] [[&short=v &extra=info]]").
			Puts(
				vals.MakeList(
					vals.MakeMap(
						"spec", vals.MakeMap("short", "v", "extra", "info"),
						"long", false, "arg", "")),
				vals.MakeList("foo")),

		// spec with &arg-required
		That("flag:parse-getopt [-p 80 foo] [[&short=p &arg-required]]").
			Puts(
				vals.MakeList(
					vals.MakeMap(
						"spec", vals.MakeMap("short", "p", "arg-required", true),
						"long", false, "arg", "80")),
				vals.MakeList("foo")),
		// spec with &arg-optional, with argument
		That("flag:parse-getopt [-i.bak foo] [[&short=i &arg-optional]]").
			Puts(
				vals.MakeList(
					vals.MakeMap(
						"spec", vals.MakeMap("short", "i", "arg-optional", true),
						"long", false, "arg", ".bak")),
				vals.MakeList("foo")),
		// spec with &arg-optional, without argument
		That("flag:parse-getopt [-i foo] [[&short=i &arg-optional]]").
			Puts(
				vals.MakeList(
					vals.MakeMap(
						"spec", vals.MakeMap("short", "i", "arg-optional", true),
						"long", false, "arg", "")),
				vals.MakeList("foo")),

		// &stop-after-double-dash on (default)
		That("flag:parse-getopt [-- -v] [[&short=v]]").
			Puts(vals.EmptyList, vals.MakeList("-v")),
		// &stop-after-double-dash off
		That("flag:parse-getopt [-- -v] [[&short=v]] &stop-after-double-dash=$false").
			Puts(vals.MakeList(vFlag), vals.MakeList("--")),
		// &stop-before-non-flag off (default)
		That("flag:parse-getopt [foo -v] [[&short=v]]").
			Puts(vals.MakeList(vFlag), vals.MakeList("foo")),
		// &stop-before-non-flag on
		That("flag:parse-getopt [foo -v] [[&short=v]] &stop-before-non-flag").
			Puts(vals.EmptyList, vals.MakeList("foo", "-v")),
		// &long-only off (default)
		That("flag:parse-getopt [-verbose] [[&long=verbose]]").
			Throws(ErrorWithMessage("unknown option -v")),
		// &long-only on
		That("flag:parse-getopt [-verbose] [[&long=verbose]] &long-only").
			Puts(
				vals.MakeList(
					vals.MakeMap(
						"spec", vals.MakeMap("long", "verbose"),
						"long", true, "arg", "")),
				vals.EmptyList),

		// None of &short and &long
		That("flag:parse-getopt [] [[&]]").Throws(errShortLong),
		// Both &arg-required and &arg-optional
		That("flag:parse-getopt [] [[&short=x &arg-optional &arg-required]]").
			Throws(errArgRequiredArgOptional),

		// Flag parsing error
		That("flag:parse-getopt [-x] []").Throws(ErrorWithMessage("unknown option -x")),
		// Bad argument list
		That("flag:parse-getopt [(num 0)] []").
			Throws(ErrorWithMessage("wrong type: need string, got number")),
		// Bad spec list
		That("flag:parse-getopt [] [(num 0)]").
			Throws(ErrorWithMessage("wrong type: need !!hashmap.Map, got number")),
	)
}

func setup(ev *eval.Evaler) { ev.ExtendGlobal(eval.BuildNs().AddNs("flag", Ns)) }

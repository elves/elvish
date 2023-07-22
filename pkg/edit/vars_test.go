package edit

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

func TestAddVar(t *testing.T) {
	TestWithEvalerSetup(t, func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddGoFn("add-var", addVar))
	},
		That("add-var foo bar").Then("put $foo").Puts("bar"),

		// Qualified name
		That("add-var a:b ''").Throws(
			errs.BadValue{
				What:  "name argument to edit:add-var",
				Valid: "unqualified variable name", Actual: "a:b"}),
		// Bad type
		That("add-var a~ ''").Throws(ErrorWithType(vals.WrongType{})),
	)
}

func TestDelVar(t *testing.T) {
	TestWithEvalerSetup(t, func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddGoFn("del-var", delVar))
	},
		That("var foo = bar").Then("del-var foo").Then("put $foo").
			DoesNotCompile("variable $foo not found"),
		// Deleting a non-existent variable is not an error
		That("del-var foo").DoesNothing(),

		// Qualified name
		That("del-var a:b").Throws(
			errs.BadValue{
				What:  "name argument to edit:del-var",
				Valid: "unqualified variable name", Actual: "a:b"}),
	)
}

func TestAddVars(t *testing.T) {
	TestWithEvalerSetup(t, func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddGoFn("add-vars", addVars))
	},
		That("add-vars [&foo=bar]").Then("put $foo").Puts("bar"),
		That("add-vars [&a=A &b=B]").Then("put $a $b").Puts("A", "B"),

		// Non-string key
		That("add-vars [&[]='']").Throws(
			errs.BadValue{
				What:  "key of argument to edit:add-vars",
				Valid: "string", Actual: "list"}),

		// Qualified name
		That("add-vars [&a:b='']").Throws(
			errs.BadValue{
				What:  "key of argument to edit:add-vars",
				Valid: "unqualified variable name", Actual: "a:b"}),
		// Bad type
		That("add-vars [&a~='']").Throws(ErrorWithType(vals.WrongType{})),
	)
}

func TestDelVars(t *testing.T) {
	TestWithEvalerSetup(t, func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddGoFn("del-vars", delVars))
	},
		That("var a b c").Then("del-vars [a b]").Then("put $a").
			DoesNotCompile("variable $a not found"),
		That("var a b c").Then("del-vars [a b]").Then("put $b").
			DoesNotCompile("variable $b not found"),
		That("var a b c").Then("del-vars [a b]").Then("put $c").Puts(nil),

		// Non-string key
		That("del-vars [[]]").Throws(
			errs.BadValue{
				What:  "element of argument to edit:del-vars",
				Valid: "string", Actual: "list"}),

		// Qualified name
		That("del-vars [a:b]").Throws(
			errs.BadValue{
				What:  "element of argument to edit:del-vars",
				Valid: "unqualified variable name", Actual: "a:b"}),
	)
}

package edit

import (
	"errors"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

func TestAddVar(t *testing.T) {
	TestWithSetup(t, func(ev *eval.Evaler) {
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
	TestWithSetup(t, func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddGoFns(map[string]any{
			"add-var": addVar,
			"del-var": delVar,
		}))
	},
		That("del-var foo").Throws(errors.New("no variable $foo")),
		That("add-var foo bar").Then("del-var foo").Then("put $foo").DoesNotCompile("variable $foo not found"),
		That("add-var foo bar").Then("del-var foo").Then("del-var foo").Throws(errors.New("no variable $foo")),

		// Qualified name
		That("del-var a:b").Throws(
			errs.BadValue{
				What:  "name argument to edit:del-var",
				Valid: "unqualified variable name", Actual: "a:b"}),
	)
}

func TestAddVars(t *testing.T) {
	TestWithSetup(t, func(ev *eval.Evaler) {
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
	TestWithSetup(t, func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddGoFn("add-vars", addVars))
		ev.ExtendGlobal(eval.BuildNs().AddGoFn("del-vars", delVars))
	},
		That("add-vars [&foo=bar]").Then("del-vars [foo]").Then("put $foo").DoesNotCompile("variable $foo not found"),
		That("add-vars [&a=A &b=B &c=C]").Then("del-vars [a b]").Then("put $c").Puts("C").Then("del-vars [c]").Then("put $c").DoesNotCompile("variable $c not found"),

		// Non-string key
		That("del-vars [[]]").Throws(
			errs.BadValue{
				What:  "name of argument to edit:del-vars",
				Valid: "string", Actual: "list"}),

		// Qualified name
		That("del-vars [a:b]").Throws(
			errs.BadValue{
				What:  "name of argument to edit:del-vars",
				Valid: "unqualified variable name", Actual: "a:b"}),
	)
}

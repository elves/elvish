package edit

import (
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
		That("add-var a~ ''").Throws(vals.WrongType{WantKind: "fn", GotKind: "string"}),
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
		That("add-vars [&a~='']").Throws(vals.WrongType{WantKind: "fn", GotKind: "string"}),
	)
}

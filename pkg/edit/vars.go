package edit

import (
	"strings"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

func initVarsAPI(ed *Editor, nb eval.NsBuilder) {
	nb.AddGoFns(map[string]interface{}{
		"add-var":  addVar,
		"add-vars": addVars,
	})
}

//elvdoc:fn add-var
//
// ```elvish
// edit:add-var $name $init
// ```
//
// Declares a new variable in the REPL with an initial value. The new variable
// becomes available during the next REPL cycle.
//
// Equivalent to running `var $name = $init` at the REPL, but `$name` can be
// dynamic.
//
// This is most useful for modules to modify the REPL namespace. Example:
//
// ```elvish-transcript
// ~> cat .config/elvish/lib/a.elv
// for i [(range 10)] {
//   edit:add-var foo$i $i
// }
// ~> use a
// ~> put $foo1 $foo2
// ▶ (num 1)
// ▶ (num 2)
// ```
//
// Note that if you pass a variable as the `$init` argument, `edit:add-var`
// doesn't add the variable "itself" to the REPL namespace, it creates a new
// variable, only using the existing variable's value:
//
// ```elvish-transcript
// ~> cat .config/elvish/lib/b.elv
// var foo = foo
// edit:add-var foo $foo
// ~> use b
// ~> put $foo
// ▶ foo
// ~> set foo = bar
// ~> echo $b:foo
// foo
// ```

func addVar(fm *eval.Frame, name string, val interface{}) error {
	if !isUnqualified(name) {
		return errs.BadValue{
			What:  "name argument to edit:add-var",
			Valid: "unqualified variable name", Actual: name}
	}
	variable := eval.MakeVarFromName(name)
	err := variable.Set(val)
	if err != nil {
		return err
	}
	fm.Evaler.ExtendGlobal(eval.BuildNs().AddVar(name, vars.FromInit(val)))
	return nil
}

//elvdoc:fn add-vars
//
// ```elvish
// edit:add-vars $map
// ```
//
// Takes a map from strings to arbitrary values. Equivalent to calling
// `edit:add-var` for each key-value pair in the map.

func addVars(fm *eval.Frame, m vals.Map) error {
	nb := eval.BuildNs()
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, val := it.Elem()
		name, ok := k.(string)
		if !ok {
			return errs.BadValue{
				What:  "key of argument to edit:add-vars",
				Valid: "string", Actual: vals.Kind(k)}
		}
		if !isUnqualified(name) {
			return errs.BadValue{
				What:  "key of argument to edit:add-vars",
				Valid: "unqualified variable name", Actual: name}
		}
		variable := eval.MakeVarFromName(name)
		err := variable.Set(val)
		if err != nil {
			return err
		}
		nb.AddVar(name, variable)
	}
	fm.Evaler.ExtendGlobal(nb)
	return nil
}

func isUnqualified(name string) bool {
	i := strings.IndexByte(name, ':')
	return i == -1 || i == len(name)-1
}

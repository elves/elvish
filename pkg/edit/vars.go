package edit

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

func initVarsAPI(ed *Editor, nb eval.NsBuilder) {
	nb.AddGoFns("<edit>:", map[string]interface{}{
		"add-var":  addVar,
		"add-vars": addVars,
	})
}

//elvdoc:fn add-var
//
// ```elvish
// edit:add-var $name $value
// ```
//
// Declares a new variable in the REPL. The new variable becomes available
// during the next REPL cycle.
//
// Equivalent to running `var $name = $value` at the REPL, but `$name` can be
// dynamic.
//
// Example:
//
// ```elvish-transcript
// ~> edit:add-var foo bar
// ~> put $foo
// ▶ bar
// ```

func addVar(fm *eval.Frame, name string, val interface{}) error {
	if !eval.IsUnqualified(name) {
		return errs.BadValue{
			What:  "name argument to edit:add-var",
			Valid: "unqualified variable name", Actual: name}
	}
	variable := eval.MakeVarFromName(name)
	err := variable.Set(val)
	if err != nil {
		return err
	}
	fm.Evaler.AddGlobal(eval.NsBuilder{name: vars.FromInit(val)}.Ns())
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
	nb := eval.NsBuilder{}
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, val := it.Elem()
		name, ok := k.(string)
		if !ok {
			return errs.BadValue{
				What:  "key of argument to edit:add-vars",
				Valid: "string", Actual: vals.Kind(k)}
		}
		if !eval.IsUnqualified(name) {
			return errs.BadValue{
				What:  "key of argument to edit:add-vars",
				Valid: "unqualified variable name", Actual: name}
		}
		variable := eval.MakeVarFromName(name)
		err := variable.Set(val)
		if err != nil {
			return err
		}
		nb[name] = variable
	}
	fm.Evaler.AddGlobal(nb.Ns())
	return nil
}

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
// Defines a new variable in the interactive REPL with an initial value. The new variable becomes
// available during the next REPL cycle.
//
// Equivalent to running `var $name = $init` at a REPL prompt, but `$name` can be
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
// Note that if you use a variable as the `$init` argument, `edit:add-var`
// doesn't add the variable "itself" to the REPL namespace. The variable in the
// REPL namespace will have the initial value set to the variable's value, but
// it is not an alias of the original variable:
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
//
// One common use of this command is to put the definitions of functions intended for REPL use in a
// module instead of your [`rc.elv`](command.html#rc-file). For example, if you want to define `ll`
// as `ls -l`, you can do so in your [`rc.elv`](command.html#rc-file) directly:
//
// ```elvish
// fn ll {|@a| ls -l $@a }
// ```
//
// But if you move the definition into a module (say `util.elv` in one of the
// [module search directories](command.html#module-search-directories), this
// function can only be used as `util:ll` (after `use util`). To make it usable
// directly as `ll`, you can add the following to `util.elv`:
//
// ```elvish
// edit:add-var ll~ $ll~
// ```
//
// Another use case is to add a module or function to the REPL namespace
// conditionally. For example, to only import [the `unix` module](unix.html)
// when actually running on UNIX, you can put the following in
// [`rc.elv`](command.html#rc-file):
//
// ```elvish
// use platform
// if $platform:is-unix {
//   use unix
//   edit:add-var unix: $unix:
// }
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

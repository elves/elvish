package edit

import (
	"errors"
	"strings"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

func initVarsAPI(ed *Editor, nb eval.NsBuilder) {
	nb.AddGoFns(map[string]any{
		"add-var":  addVar,
		"add-vars": addVars,
		"del-var":  delVar,
	})
}

func addVar(fm *eval.Frame, name string, val any) error {
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

func delVar(fm *eval.Frame, name string) error {
	if !isUnqualified(name) {
		return errs.BadValue{
			What:  "name argument to edit:del-var",
			Valid: "unqualified variable name", Actual: name}
	}

	variable := fm.Evaler.Global().IndexString(name)
	if variable == nil {
		return errors.New("no variable $" + name)
	}

	fm.Evaler.DeleteGlobal(name)
	return nil
}

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

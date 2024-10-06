package etk

import (
	"fmt"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/ui"
)

type textViewOpts struct{ DotBefore int }

func (*textViewOpts) SetDefaultOptions() {}

var Ns = eval.BuildNsNamed("etk").
	AddVars(map[string]vars.Var{
		// Reaction values
		"unused":     vars.NewReadOnly(Unused),
		"consumed":   vars.NewReadOnly(Consumed),
		"finish":     vars.NewReadOnly(Finish),
		"finish-eof": vars.NewReadOnly(FinishEOF),
		// Builtin widgets
		"textarea": vars.NewReadOnly(TextArea),
	}).
	AddGoFns(map[string]any{
		"-text-view": func(opts textViewOpts, v any) View {
			return TextView(opts.DotBefore, ui.T(vals.ToString(v)))
		},
		"-key-event": func(s string) (term.Event, error) {
			k, err := ui.ParseKey(s)
			if err != nil {
				return nil, err
			}
			return term.KeyEvent(k), nil
		},
		"with-init": func(fm *eval.Frame, compAny any, inits vals.Map) (Comp, error) {
			// TODO: Integrate the parsing into vals.ScanToGo
			comp, err := scanComp(fm, compAny)
			if err != nil {
				return nil, err
			}
			initArgs, err := convertInits(inits)
			if err != nil {
				return nil, err
			}
			return WithInit(comp, initArgs...), nil
		},
		"run": func(fm *eval.Frame, compAny any) error {
			// TODO: Integrate the parsing into vals.ScanToGo
			comp, err := scanComp(fm, compAny)
			if err != nil {
				return err
			}
			// TODO: Maybe should use subframe of fm?
			_, err = Run(cli.NewTTY(fm.InputFile(), fm.Port(1).File), fm, comp)
			return err
		},
	}).
	Ns()

func scanComp(fm *eval.Frame, v any) (Comp, error) {
	switch v := v.(type) {
	case Comp:
		return v, nil
	case eval.Callable:
		return scanCompFromFn(fm, v), nil
	default:
		// TODO: Proper error reporting
		return nil, fmt.Errorf("need function as comp, got %#v; traceback: %v", v, fm.TraceBack())
		// return nil, errs.BadValue{What: "comp", Valid: "function", Actual: fmt.Sprint(v)}
	}
}

type compOut struct {
	View  View
	React eval.Callable
}

type viewReact struct {
	View  View
	React React
}

type vboxOpts struct{ Focus int }

func (*vboxOpts) SetDefaultOptions() {}

func scanCompFromFn(fm *eval.Frame, fn eval.Callable) Comp {
	return func(c Context) (View, React) {
		subcomps := map[string]viewReact{}
		var elvishCtx = eval.BuildNs().AddVars(map[string]vars.Var{
			"state": stateSubTreeVar(c),
		}).AddGoFns(map[string]any{
			"state": func(name string, _eq string, init any) {
				State(c, name, init)
			},
			"subcomp": func(name string, _eq string, compAny any) error {
				// TODO: Is use of fm correct here?
				comp, err := scanComp(fm, compAny)
				if err != nil {
					return err
				}
				v, r := c.Subcomp(name, comp)
				subcomps[name] = viewReact{v, r}
				return nil
			},
			"vbox": func(opts vboxOpts, compNames ...string) View {
				views := make([]View, len(compNames))
				for i, compName := range compNames {
					views[i] = subcomps[compName].View
				}
				return VBoxView(opts.Focus, views...)
			},
			"pass": func(compName string, ev term.Event) Reaction {
				// TODO: Error if comp doesn't exist
				return subcomps[compName].React(ev)
			},
		}).Ns()
		p1, getOut, err := eval.ValueCapturePort()
		if err != nil {
			return errElement(err)
		}
		// TODO: How to call this properly?
		err = fm.Evaler.Call(fn, eval.CallCfg{Args: []any{elvishCtx}},
			eval.EvalCfg{Ports: []*eval.Port{nil, p1, nil}})
		if err != nil {
			return errElement(err)
		}
		outs := getOut()
		if len(outs) != 1 {
			return errElement(fmt.Errorf("should only have one output"))
		}
		var out compOut
		err = vals.ScanToGo(outs[0], &out)
		if err != nil {
			return errElement(fmt.Errorf("output should be map with view and react"))
		}
		// TODO: Handle scan error
		return out.View, must.OK1(ScanToGo[React](out.React, fm))
	}
}

type stateSubTreeVar Context

func (v stateSubTreeVar) Get() any {
	return getPath(v.g.state, v.path)
}

func (v stateSubTreeVar) Set(val any) error {
	valMap, ok := val.(vals.Map)
	if !ok {
		return fmt.Errorf("must be map")
	}
	v.g.state = assocPath(v.g.state, v.path, valMap)
	return nil
}

func convertInits(m vals.Map) ([]any, error) {
	var args []any
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		name, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("key should be string")
		}
		args = append(args, name, v)
	}
	return args, nil
}

func errElement(err error) (View, React) {
	return TextView(1, ui.T(err.Error(), ui.FgRed)),
		func(term.Event) Reaction { return Unused }
}

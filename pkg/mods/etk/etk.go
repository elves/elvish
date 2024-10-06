package etk

import (
	"fmt"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
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
		// etk.Reaction values
		"unused":     vars.NewReadOnly(etk.Unused),
		"consumed":   vars.NewReadOnly(etk.Consumed),
		"finish":     vars.NewReadOnly(etk.Finish),
		"finish-eof": vars.NewReadOnly(etk.FinishEOF),
		// Builtin widgets
		"textarea": vars.NewReadOnly(comps.TextArea),
	}).
	AddGoFns(map[string]any{
		/*
			"-text-view": func(opts textViewOpts, v any) etk.View {
				return etk.TextView(opts.DotBefore, ui.T(vals.ToString(v)))
			},
		*/
		"-key-event": func(s string) (term.Event, error) {
			k, err := ui.ParseKey(s)
			if err != nil {
				return nil, err
			}
			return term.KeyEvent(k), nil
		},
		"with-init": func(fm *eval.Frame, compAny any, inits vals.Map) (etk.Comp, error) {
			// TODO: Integrate the parsing into vals.ScanToGo
			comp, err := scanComp(fm, compAny)
			if err != nil {
				return nil, err
			}
			initArgs, err := convertInits(inits)
			if err != nil {
				return nil, err
			}
			return etk.WithInit(comp, initArgs...), nil
		},
		"run": func(fm *eval.Frame, compAny any) error {
			// TODO: Integrate the parsing into vals.ScanToGo
			comp, err := scanComp(fm, compAny)
			if err != nil {
				return err
			}
			// TODO: Maybe should use subframe of fm?
			_, err = etk.Run(comp, etk.RunCfg{
				TTY: cli.NewTTY(fm.InputFile(), fm.Port(1).File), Frame: fm})
			return err
		},
	}).
	Ns()

func scanComp(fm *eval.Frame, v any) (etk.Comp, error) {
	switch v := v.(type) {
	case etk.Comp:
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
	View  etk.View
	React eval.Callable
}

type viewReact struct {
	View  etk.View
	React etk.React
}

type vboxOpts struct{ Focus int }

func (*vboxOpts) SetDefaultOptions() {}

func scanCompFromFn(fm *eval.Frame, fn eval.Callable) etk.Comp {
	return func(c etk.Context) (etk.View, etk.React) {
		subcomps := map[string]viewReact{}
		var elvishCtx = eval.BuildNs().AddVars(map[string]vars.Var{
			"state": etk.StateSubTreeVar(c),
		}).AddGoFns(map[string]any{
			"state": func(name string, _eq string, init any) {
				etk.State(c, name, init)
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
			/*
				"vbox": func(opts vboxOpts, compNames ...string) etk.View {
					views := make([]etk.View, len(compNames))
					for i, compName := range compNames {
						views[i] = subcomps[compName].View
					}
					return etk.VBoxView(opts.Focus, views...)
				},
			*/
			"pass": func(compName string, ev term.Event) etk.Reaction {
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
		return out.View, must.OK1(etk.ScanToGo[etk.React](out.React, fm))
	}
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

func errElement(err error) (etk.View, etk.React) {
	return etk.Text(ui.T(err.Error(), ui.FgRed), etk.DotHere),
		func(term.Event) etk.Reaction { return etk.Unused }
}

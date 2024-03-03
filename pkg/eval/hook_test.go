package eval_test

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
)

func TestCallHook(t *testing.T) {
	v := vals.EmptyList
	hook := vars.FromPtr(&v)
	ev := eval.NewEvaler()
	ev.ExtendGlobal(eval.BuildNs().AddVar("test-hook", hook))

	ports := eval.DummyPorts[:]
	p, collect, err := eval.CapturePort()
	if err != nil {
		t.Error(err)
	}
	ports[1] = p

	evalCfg := eval.EvalCfg{Ports: ports}

	code := `set test-hook = [{|| put $true }]`
	err = ev.Eval(parse.Source{Name: "[test]", Code: code}, evalCfg)
	if err != nil {
		t.Error(err)
	}
	eval.CallHook(ev, &evalCfg, "test-hook", hook.Get().(vals.List))

	vs, _ := collect()
	if len(vs) != 1 {
		t.Error(len(vs))
	}

	if v, ok := vs[0].(bool); !ok || !v {
		t.Error(vs)
	}
}

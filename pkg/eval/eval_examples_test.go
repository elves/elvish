package eval_test

import (
	"fmt"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
)

func ExampleEvaler_Eval_usingPortsFromStdFiles() {
	ev := eval.NewEvaler()
	// These ports are connected to the process's stdin, stdout and stderr. The
	// "> " part is the prefix to use for value outputs.
	ports, cleanup := eval.PortsFromStdFiles("> ")
	defer cleanup()

	ev.Eval(
		parse.Source{Name: "example 1", Code: "echo Hello Elvish!"},
		eval.EvalCfg{Ports: ports})

	// Value outputs are written with the prefix we specified earlier
	ev.Eval(
		parse.Source{Name: "example 2", Code: "put [&foo=bar]"},
		eval.EvalCfg{Ports: ports})

	// Output:
	// Hello Elvish!
	// > [&foo=bar]
}

func ExampleEvaler_Eval_capturingValueOutputs() {
	ev := eval.NewEvaler()
	// The stdout port captures all values written to it, which can be retrieved
	// with the returned get function.
	stdout, get, err := eval.ValueCapturePort()
	if err != nil {
		panic(err)
	}

	ev.Eval(
		parse.Source{Name: "example 1", Code: "put [&foo=bar] [a b] data"},
		eval.EvalCfg{Ports: []*eval.Port{eval.DummyInputPort, stdout, eval.DummyOutputPort}})

	values := get()
	for i, value := range values {
		fmt.Printf("#%d: %s: %s\n", i, vals.Kind(value), vals.ReprPlain(value))
	}

	// Output:
	// #0: map: [&foo=bar]
	// #1: list: [a b]
	// #2: string: data
}

func ExampleEvaler_Eval_inspectingGlobal() {
	ev := eval.NewEvaler()
	ev.Eval(
		parse.Source{Name: "example 1", Code: "var map = [&foo=bar]"},
		// Omitting the ports connects all of them to "dummy" IO ports.
		eval.EvalCfg{})

	m, ok := ev.Global().Index("map")
	if !ok {
		fmt.Println("$map not found")
	}
	fmt.Printf("$map: %s: %s\n", vals.Kind(m), vals.ReprPlain(m))

	// Output:
	// $map: map: [&foo=bar]
}

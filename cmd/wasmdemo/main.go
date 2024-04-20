package main

import (
	"fmt"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

const code = `
fn hello {|who|
  echo 'Hello, '$who'!'
}
hello browser
`

func main() {
	ev := eval.NewEvaler()
	ports, cleanup := eval.PortsFromStdFiles(ev.ValuePrefix())
	defer cleanup()
	err := ev.Eval(parse.Source{Code: code}, eval.EvalCfg{Ports: ports})
	if err != nil {
		fmt.Println("eval error:", err)
	}
}

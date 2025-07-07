// Example terminal apps implemented using the Etk framework.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/ui"
)

var (
	showState = flag.Bool("show-state", false, "show app state")
	maxHeight = flag.Int("max-height", 0, "max height of the app")
	justify   = flag.String("justify", "none", "how to justify app vertically")
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintln(os.Stderr, "need example name")
		os.Exit(1)
	}
	j := etk.NoJustify
	switch *justify {
	case "none":
		j = etk.NoJustify
	case "top":
		j = etk.JustifyTop
	case "center":
		j = etk.JustifyCenter
	case "bottom":
		j = etk.JustifyBottom
	default:
		fmt.Fprintln(os.Stderr, "unknown justify:", *justify)
		os.Exit(1)
	}
	example := flag.Args()[0]

	var f etk.Comp
	switch example {
	// 7GUIs
	case "counter":
		f = Counter
	case "counter-with-button":
		f = CounterWithButton
	case "temperature":
		f = Temperature
	case "flight":
		f = Flight
	case "timer":
		f = Timer
	case "crud":
		f = CRUD

	case "textarea":
		f = TextArea
	case "wizard":
		f = Wizard
	case "todo":
		f = Todo
	case "preso":
		content := must.OK1(os.ReadFile(flag.Args()[1]))
		pages := parsePreso(string(content))
		f = etk.ModComp(Preso, etk.InitState("pages", pages))
	case "datanav":
		f = etk.ModComp(comps.HierNav,
			etk.InitState("hier", dataHier{hierData}),
			etk.InitState("path", []string{"bin"}))
	case "fsnav":
		f = etk.ModComp(comps.HierNav, etk.InitState("hier", fsHier{}))
	case "life":
		f = etk.ModComp(Life,
			etk.InitState("name", "pentadecathlon"),
			etk.InitState("history", []Board{pentadecathlon}))
	case "styledown":
		f = Styledown
	default:
		fmt.Fprintln(os.Stderr, "unknown example:", example)
		os.Exit(1)
	}
	f = QuitWithEsc(f)
	if *showState {
		f = etk.ModComp(ShowState, etk.InitState("inner-comp", f))
	}
	etk.Run(f, etk.RunCfg{
		TTY:       cli.NewTTY(os.Stdin, os.Stdout),
		Frame:     eval.NewEvaler().CallFrame("etk"),
		MaxHeight: *maxHeight, Justify: j,
	})
}

func QuitWithEsc(f etk.Comp) etk.Comp {
	return func(c etk.Context) (etk.View, etk.React) {
		v, r := f(c)
		return v, func(ev term.Event) etk.Reaction {
			reaction := r(ev)
			if reaction == etk.Unused && (ev == term.K('[', ui.Ctrl)) {
				return etk.Finish
			}
			return reaction
		}
	}
}

func ShowState(c etk.Context) (etk.View, etk.React) {
	innerView, innerReact := c.Subcomp("inner", nop)
	innerStateVar := etk.BindState(c, "inner", vals.EmptyMap)

	stateText := ui.T(strings.ReplaceAll(vals.Repr(innerStateVar.Get(), 0), "\t", " "))

	return etk.Box(`
		[inner*]
		state*`, innerView, etk.Text(stateText)), innerReact
}

func nop(c etk.Context) (etk.View, etk.React) {
	return etk.EmptyView{}, func(term.Event) etk.Reaction { return etk.Unused }
}

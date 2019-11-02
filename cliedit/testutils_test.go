package cliedit

import (
	"fmt"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

var styles = map[rune]string{
	'-': "underlined",
	'm': "bold lightgray bg-magenta", // mode line
	'#': "inverse",
	'g': "green",   // good
	'b': "red",     // bad
	'v': "magenta", // variables
	'e': "bg-red",  // error
}

const (
	testTTYWidth  = 40
	testTTYHeight = 24
)

func bb() *ui.BufferBuilder { return ui.NewBufferBuilder(testTTYWidth) }

func prepare() (cli.App, cli.TTYCtrl, eval.Ns, *eval.Evaler) {
	appSpec, ns, ev := preparePreApp()
	app, ttyCtrl := prepareApp(appSpec, ns, ev)
	return app, ttyCtrl, ns, ev
}

func preparePreApp() (cli.AppSpec, eval.Ns, *eval.Evaler) {
	return cli.AppSpec{}, eval.Ns{}, eval.NewEvaler()
}

func prepareApp(spec cli.AppSpec, ns eval.Ns, ev *eval.Evaler) (cli.App, cli.TTYCtrl) {
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(24, 40)
	spec.TTY = tty
	app := cli.NewApp(spec)
	ev.InstallModule("edit", ns)
	evalf(ev, "use edit")
	return app, ttyCtrl
}

func run(app cli.App) func() {
	codeCh, _ := cli.ReadCodeAsync(app)
	return func() {
		app.CommitEOF()
		<-codeCh
	}
}

func evalf(ev *eval.Evaler, format string, args ...interface{}) {
	code := fmt.Sprintf(format, args...)
	// TODO: Should use a difference source type
	err := ev.EvalSourceInTTY(eval.NewInteractiveSource(code))
	if err != nil {
		panic(err)
	}
}

func getNs(ns eval.Ns, name string) eval.Ns {
	return ns[name+eval.NsSuffix].Get().(eval.Ns)
}

func getFn(ns eval.Ns, name string) eval.Callable {
	return ns[name+eval.FnSuffix].Get().(eval.Callable)
}

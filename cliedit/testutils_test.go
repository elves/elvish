package cliedit

import (
	"fmt"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/eval"
)

func prepare() (cli.App, cli.TTYCtrl, eval.Ns, *eval.Evaler) {
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(24, 40)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	ns := eval.Ns{}
	ev := eval.NewEvaler()
	ev.InstallModule("edit", ns)
	evalf(ev, "use edit")
	return app, ttyCtrl, ns, ev
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

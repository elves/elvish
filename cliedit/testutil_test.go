package cliedit

import (
	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/cli/term"
)

func setup() (*clicore.App, clicore.TTYCtrl, clicore.SignalSourceCtrl) {
	tty, ttyControl := clicore.NewFakeTTY()
	sigs, sigsCtrl := clicore.NewFakeSignalSource()
	app := clicore.NewApp(tty, sigs)
	return app, ttyControl, sigsCtrl
}

func cleanup(tty clicore.TTYCtrl, codeCh <-chan string) {
	// Causes BasicMode to quit
	tty.Inject(term.K('\n'))
	// Wait until ReadCode has finished execution
	<-codeCh
}

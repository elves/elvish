package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/eval"
)

//elvdoc:fn binding-map
//
// Converts a normal map into a binding map.

var makeBindingMap = eddefs.MakeBindingMap

//elvdoc:fn commit-code
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read. Internally, this works by raising a special exception.

//elvdoc:fn commit-eof
//
// Causes the Elvish REPL to terminate. Internally, this works by raising a
// special exception.

func initMiscBuiltins(app *cli.App, ns eval.Ns) {
	ns.AddGoFns("<edit>", map[string]interface{}{
		"binding-map": makeBindingMap,
		"commit-code": app.CommitCode,
		"commit-eof":  app.CommitEOF,
	})
}

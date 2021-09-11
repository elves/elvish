// Package daemon implements a service for mediating access to the data store,
// and its client.
//
// Most RPCs exposed by the service correspond to the methods of Store in the
// store package and are not documented here.
package daemon

import (
	"os"

	"src.elv.sh/pkg/logutil"
	"src.elv.sh/pkg/prog"
)

var logger = logutil.GetLogger("[daemon] ")

// Program is the daemon subprogram.
var Program prog.Program = program{}

type program struct{}

func (program) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	if !f.Daemon {
		return prog.ErrNotSuitable
	}
	if len(args) > 0 {
		return prog.BadUsage("arguments are not allowed with -daemon")
	}
	setUmaskForDaemon()
	Serve(f.Sock, f.DB)
	return nil
}

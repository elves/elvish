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

// Version is the API version. It should be bumped any time the API changes.
const Version = -93

// Program is the daemon subprogram.
var Program prog.Program = program{}

type program struct{}

func (program) ShouldRun(f *prog.Flags) bool { return f.Daemon }

func (program) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	if len(args) > 0 {
		return prog.BadUsage("arguments are not allowed with -daemon")
	}
	setUmaskForDaemon()
	Serve(f.Sock, f.DB)
	return nil
}

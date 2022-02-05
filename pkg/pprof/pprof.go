// Package pprof adds profiling support to the Elvish program.
package pprof

import (
	"fmt"
	"os"
	"runtime/pprof"

	"src.elv.sh/pkg/prog"
)

// Program adds support for the -cpuprofile flag.
type Program struct {
	cpuProfile string
}

func (p *Program) RegisterFlags(f *prog.FlagSet) {
	f.StringVar(&p.cpuProfile, "cpuprofile", "", "write CPU profile to file")
}

func (p *Program) Run(fds [3]*os.File, _ []string) error {
	if p.cpuProfile != "" {
		f, err := os.Create(p.cpuProfile)
		if err != nil {
			fmt.Fprintln(fds[2], "Warning: cannot create CPU profile:", err)
			fmt.Fprintln(fds[2], "Continuing without CPU profiling.")
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}
	return prog.ErrNextProgram
}

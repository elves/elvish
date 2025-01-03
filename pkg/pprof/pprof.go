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
	cpuProfile    string
	allocsProfile string
}

func (p *Program) RegisterFlags(f *prog.FlagSet) {
	f.StringVar(&p.cpuProfile, "cpuprofile", "", "write CPU profile to file")
	f.StringVar(&p.allocsProfile, "allocsprofile", "", "write memory allocation profile to file")
}

func (p *Program) Run(fds [3]*os.File, _ []string) error {
	var cleanups []func([3]*os.File)
	if p.cpuProfile != "" {
		f, err := os.Create(p.cpuProfile)
		if err != nil {
			fmt.Fprintln(fds[2], "Warning: cannot create CPU profile:", err)
			fmt.Fprintln(fds[2], "Continuing without CPU profiling.")
		} else {
			pprof.StartCPUProfile(f)
			cleanups = append(cleanups, func([3]*os.File) {
				pprof.StopCPUProfile()
				f.Close()
			})
		}
	}
	if p.allocsProfile != "" {
		f, err := os.Create(p.allocsProfile)
		if err != nil {
			fmt.Fprintln(fds[2], "Warning: cannot create memory allocation profile:", err)
			fmt.Fprintln(fds[2], "Continuing without memory allocation profiling.")
		} else {
			cleanups = append(cleanups, func([3]*os.File) {
				pprof.Lookup("allocs").WriteTo(f, 0)
				f.Close()
			})
		}
	}
	return prog.NextProgram(cleanups...)
}

// Elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/elves/elvish/shell"
	"github.com/elves/elvish/util"
)

var Logger = util.GetLogger("[main] ")

var (
	log        = flag.String("log", "", "a file to write debug log to")
	dbname     = flag.String("db", "", "path to the database")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	help       = flag.Bool("help", false, "show usage help and quit")
	cmd        = flag.Bool("c", false, "take first argument as a command to execute")
)

func usage() {
	fmt.Println("usage: elvish [flags] [script]")
	fmt.Println("flags:")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if *help {
		usage()
		os.Exit(0)
	}

	if len(args) > 1 {
		usage()
		os.Exit(2)
	}

	if *log != "" {
		err := util.SetOutputFile(*log)
		if err != nil {
			fmt.Println(err)
		}
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			Logger.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	sh := shell.NewShell(*dbname, *cmd)

	var arg *string
	if len(args) == 1 {
		arg = &args[0]
	}

	os.Exit(sh.Main(arg))
}

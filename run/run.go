// Package run is the entry point of elvish.
package run

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/elves/elvish/edit"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/stub"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var Logger = util.GetLogger("[main] ")

var (
	log        = flag.String("log", "", "a file to write debug log to")
	dbname     = flag.String("db", "", "path to the database")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	help       = flag.Bool("help", false, "show usage help and quit")
)

func usage() {
	fmt.Println("usage: elvish [flags] [script]")
	fmt.Println("flags:")
	flag.PrintDefaults()
}

// Main is the entry point of elvish.
func Main() {
	defer rescue()

	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) > 1 {
		usage()
		os.Exit(2)
	}

	if *help {
		usage()
		os.Exit(0)
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

	handleHupAndQuit()
	logSignals()

	ev, st := newEvalerAndStore()
	defer func() {
		err := st.Close()
		if err != nil {
			fmt.Println("failed to close database:", err)
		}
	}()

	stub, err := stub.NewStub(os.Stderr)
	if err != nil {
		fmt.Println("failed to spawn stub:", err)
	} else {
		ev.Stub = stub
		go func() {
			<-stub.State()
			fmt.Println("stub has died")
		}()
	}

	if len(args) == 1 {
		script(ev, args[0])
	} else if !sys.IsATTY(0) {
		script(ev, "/dev/stdin")
	} else {
		interact(ev, st)
	}
}

func rescue() {
	r := recover()
	if r != nil {
		println()
		fmt.Println(r)
		print(sys.DumpStack())
		println("\nexecing recovery shell /bin/sh")
		syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
	}
}

func script(ev *eval.Evaler, fname string) {
	err := ev.Source(fname)
	if err != nil {
		printError(err)
		os.Exit(1)
	}
}

func interact(ev *eval.Evaler, st *store.Store) {
	// Build Editor.
	sigch := make(chan os.Signal)
	signal.Notify(sigch)
	ed := edit.NewEditor(os.Stdin, sigch, ev, st)

	// Source rc.elv.
	datadir, err := store.EnsureDataDir()
	printError(err)
	if err == nil {
		// XXX
		err := ev.Source(datadir + "/rc.elv")
		if err != nil && !os.IsNotExist(err) {
			printError(err)
		}
	}

	// Build readLine function.
	readLine := func() (string, error) {
		return ed.ReadLine()
	}

	cooldown := time.Second
	usingBasic := false
	cmdNum := 0

	for {
		cmdNum++
		// name := fmt.Sprintf("<tty %d>", cmdNum)

		line, err := readLine()

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Editor error:", err)
			if !usingBasic {
				fmt.Println("Falling back to basic line editor")
				readLine = basicReadLine
				usingBasic = true
			} else {
				fmt.Println("Don't know what to do, pid is", os.Getpid())
				fmt.Println("Restarting editor in", cooldown)
				time.Sleep(cooldown)
				if cooldown < time.Minute {
					cooldown *= 2
				}
			}
			continue
		}

		// No error; reset cooldown.
		cooldown = time.Second

		n, err := parse.Parse(line)
		printError(err)

		if err == nil {
			err := ev.EvalInteractive(line, n)
			printError(err)
		}
	}
}

func basicReadLine() (string, error) {
	stdin := bufio.NewReaderSize(os.Stdin, 0)
	return stdin.ReadString('\n')
}

func logSignals() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	go func() {
		for sig := range sigs {
			Logger.Println("signal", sig)
		}
	}()
}

func handleHupAndQuit() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		for sig := range sigs {
			fmt.Print(sys.DumpStack())
			if sig == syscall.SIGQUIT {
				os.Exit(3)
			}
		}
	}()
}

func newEvalerAndStore() (*eval.Evaler, *store.Store) {
	dataDir, err := store.EnsureDataDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: cannot create data dir ~/.elvish")
	}

	var st *store.Store
	if err == nil {
		db := *dbname
		if db == "" {
			db = dataDir + "/db"
		}
		st, err = store.NewStore(db)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: cannot connect to store:", err)
		}
	}

	return eval.NewEvaler(st), st
}

func printError(err error) {
	if err == nil {
		return
	}
	switch err := err.(type) {
	case *util.Errors:
		for _, e := range err.Errors {
			printError(e)
		}
	default:
		eval.PprintError(err)
		fmt.Println()
	}
}

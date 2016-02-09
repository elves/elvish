// elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"syscall"

	"github.com/elves/elvish/edit"
	"github.com/elves/elvish/errutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/logutil"
	"github.com/elves/elvish/osutil"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/sys"
)

const (
	sigchSize     = 32
	outChanSize   = 32
	outChanLeader = "▶ "
)

var Logger = logutil.Discard

func usage() {
	fmt.Println("usage: elvish [flags] [script]")
	fmt.Println("flags:")
	flag.PrintDefaults()
}

var (
	debuglog = flag.String("debuglog", "", "a file to write debug log to")
	help     = flag.Bool("help", false, "show usage help and quit")
)

func main() {
	defer rescue()

	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
		os.Exit(0)
	}

	if *debuglog != "" {
		f, err := os.OpenFile(*debuglog, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		Logger = log.New(f, "[main]", log.LstdFlags)
		eval.Logger = log.New(f, "[eval] ", log.LstdFlags)
		edit.Logger = log.New(f, "[edit] ", log.LstdFlags)
	}

	go dumpstackOnQuit()
	go logSignals()

	args := flag.Args()
	switch len(args) {
	case 0:
		interact()
	case 1:
		script(args[0])
	default:
		usage()
		os.Exit(2)
	}
}

func rescue() {
	r := recover()
	if r != nil {
		print(sys.DumpStack())
		println("\nexecing recovery shell /bin/sh")
		syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
	}
}

// TODO(xiaq): Currently only the editor deals with signals.
func interact() {
	ev, st := newEvalerAndStore()
	datadir, err := store.EnsureDataDir()
	printError(err)
	if err == nil {
		// XXX
		err := ev.Source(datadir + "/rc.elv")
		if err != nil && !os.IsNotExist(err) {
			printError(err)
		}
	}

	cmdNum := 0

	username := "???"
	user, err := user.Current()
	if err == nil {
		username = user.Username
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "???"
	}
	rpromptStr := username + "@" + hostname

	sigch := make(chan os.Signal, sigchSize)
	signal.Notify(sigch)

	ed := edit.NewEditor(os.Stdin, sigch, ev, st)

	for {
		cmdNum++
		// name := fmt.Sprintf("<tty %d>", cmdNum)

		prompt := func() string {
			return osutil.Getwd() + "> "
		}
		rprompt := func() string {
			return rpromptStr
		}

		lr := ed.ReadLine(prompt, rprompt)
		// signal.Stop(sigch)

		if lr.EOF {
			break
		} else if lr.Err != nil {
			fmt.Println("Editor error:", lr.Err)
			fmt.Println("My pid is", os.Getpid())
		}

		n, err := parse.Parse(lr.Line)
		printError(err)

		if err == nil {
			err := ev.EvalInteractive(lr.Line, n)
			printError(err)
		}
	}
}

func logSignals() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	for sig := range sigs {
		Logger.Println("signal", sig)
	}
}

func dumpstackOnQuit() {
	quitSigs := make(chan os.Signal)
	signal.Notify(quitSigs, syscall.SIGQUIT)
	for range quitSigs {
		fmt.Print(sys.DumpStack())
	}
}

func script(fname string) {
	ev, _ := newEvalerAndStore()
	err := ev.Source(fname)
	if err != nil {
		printError(err)
		os.Exit(1)
	}
}

func newEvalerAndStore() (*eval.Evaler, *store.Store) {
	dataDir, err := store.EnsureDataDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: cannot create data dir ~/.elvish")
	}

	var st *store.Store
	if err == nil {
		st, err = store.NewStore(dataDir)
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
	case *errutil.ContextualError:
		fmt.Print(err.Pprint())
	case *errutil.Errors:
		for _, e := range err.Errors {
			printError(e)
		}
	default:
		eval.PprintError(err)
		fmt.Println()
	}
}

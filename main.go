// elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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

var Logger = logutil.GetLogger("[main] ")

func usage() {
	fmt.Println("usage: elvish [flags] [script]")
	fmt.Println("flags:")
	flag.PrintDefaults()
}

var (
	log    = flag.String("log", "", "a file to write debug log to")
	dbname = flag.String("db", "", "path to the database")
	help   = flag.Bool("help", false, "show usage help and quit")
)

func main() {
	defer rescue()

	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
		os.Exit(0)
	}

	if *log != "" {
		err := logutil.SetOutputFile(*log)
		if err != nil {
			fmt.Println(err)
		}
	}

	go handleQuit()
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
		println()
		fmt.Println(r)
		print(sys.DumpStack())
		println("\nexecing recovery shell /bin/sh")
		syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
	}
}

// TODO(xiaq): Currently only the editor deals with signals.
func interact() {
	ev, st := newEvalerAndStore()

	sigch := make(chan os.Signal, sigchSize)
	signal.Notify(sigch)

	ed := edit.NewEditor(os.Stdin, sigch, ev, st)

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
	prompt := func() string {
		return osutil.Getwd() + "> "
	}
	rprompt := func() string {
		return rpromptStr
	}

	readLine := func() edit.LineRead {
		return ed.ReadLine(prompt, rprompt)
	}

	usingBasic := false

	if !sys.IsATTY(0) {
		readLine = basicReadLine
		usingBasic = true
	}

	for {
		cmdNum++
		// name := fmt.Sprintf("<tty %d>", cmdNum)

		lr := readLine()
		// signal.Stop(sigch)

		if lr.EOF {
			break
		} else if lr.Err != nil {
			fmt.Println("Editor error:", lr.Err)
			if !usingBasic {
				fmt.Println("Falling back to basic line editor")
				readLine = basicReadLine
				usingBasic = true
			}
			continue
		}

		n, err := parse.Parse(lr.Line)
		printError(err)

		if err == nil {
			err := ev.EvalInteractive(lr.Line, n)
			printError(err)
		}
	}
}

func basicReadLine() edit.LineRead {
	stdin := bufio.NewReaderSize(os.Stdin, 0)
	line, err := stdin.ReadString('\n')
	if err == nil {
		return edit.LineRead{Line: line}
	} else if err == io.EOF {
		return edit.LineRead{EOF: true}
	} else {
		return edit.LineRead{Err: err}
	}
}

func logSignals() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	for sig := range sigs {
		Logger.Println("signal", sig)
	}
}

func handleQuit() {
	quitSigs := make(chan os.Signal)
	signal.Notify(quitSigs, syscall.SIGQUIT)
	<-quitSigs
	fmt.Print(sys.DumpStack())
	os.Exit(3)
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

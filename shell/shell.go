// Package shell is the entry point of elvish.
package shell

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/elves/elvish/edit"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var Logger = util.GetLogger("[shell] ")

// Shell keeps flags to the shell.
type Shell struct {
	dbname string
	cmd    bool
}

func NewShell(dbname string, cmd bool) *Shell {
	return &Shell{dbname, cmd}
}

// Main is the entry point of elvish.
func (sh *Shell) Main(arg *string) int {
	defer rescue()

	handleUsr1AndQuit()
	logSignals()

	ev, st := newEvalerAndStore(sh.dbname)
	defer func() {
		err := st.Close()
		if err != nil {
			fmt.Println("failed to close database:", err)
		}
	}()

	if arg != nil {
		if sh.cmd {
			sourceTextAndPrintError(ev, "code from -c", *arg)
		} else {
			script(ev, *arg)
		}
	} else if !sys.IsATTY(0) {
		script(ev, "/dev/stdin")
	} else {
		interact(ev, st)
	}

	return 0
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
	if !source(ev, fname, false) {
		os.Exit(1)
	}
}

func source(ev *eval.Evaler, fname string, notexistok bool) bool {
	src, err := readFileUTF8(fname)
	if err != nil {
		if notexistok && os.IsNotExist(err) {
			return true
		}
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	return sourceTextAndPrintError(ev, fname, src)
}

// sourceTextAndPrintError sources text, prints error if there is any, and
// returns whether there was no error.
func sourceTextAndPrintError(ev *eval.Evaler, name, src string) bool {
	err := ev.SourceText(name, src)
	if err != nil {
		switch err := err.(type) {
		case util.Pprinter:
			fmt.Fprintln(os.Stderr, err.Pprint(""))
		default:
			fmt.Fprintf(os.Stderr, "\033[31;1m%s\033[m", err.Error())
		}
		return false
	}
	return true
}

func readFileUTF8(fname string) (string, error) {
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(bytes) {
		return "", fmt.Errorf("%s: source is not valid UTF-8", fname)
	}
	return string(bytes), nil
}

func interact(ev *eval.Evaler, st *store.Store) {
	// Build Editor.
	sigch := make(chan os.Signal)
	signal.Notify(sigch)
	ed := edit.NewEditor(os.Stdin, sigch, ev, st)

	// Source rc.elv.
	if ev.DataDir != "" {
		source(ev, ev.DataDir+"/rc.elv", true)
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

		sourceTextAndPrintError(ev, "[interactive]", line)
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

func handleUsr1AndQuit() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGQUIT)
	go func() {
		for sig := range sigs {
			fmt.Print(sys.DumpStack())
			if sig == syscall.SIGQUIT {
				os.Exit(3)
			}
		}
	}()
}

func newEvalerAndStore(db string) (*eval.Evaler, *store.Store) {
	dataDir, err := store.EnsureDataDir()

	var st *store.Store
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: cannot create data dir ~/.elvish")
	} else {
		if db == "" {
			db = dataDir + "/db"
		}
		st, err = store.NewStore(db)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: cannot connect to store:", err)
		}
	}

	return eval.NewEvaler(st, dataDir), st
}

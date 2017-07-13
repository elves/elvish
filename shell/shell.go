// Package shell is the entry point for the terminal interface of Elvish.
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

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/edit"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[shell] ")

// Shell keeps flags to the shell.
type Shell struct {
	ev     *eval.Evaler
	daemon *api.Client
	cmd    bool
}

func NewShell(ev *eval.Evaler, daemon *api.Client, cmd bool) *Shell {
	return &Shell{ev, daemon, cmd}
}

// Run runs Elvish using the default terminal interface. It blocks until Elvish
// quites, and returns the exit code.
func (sh *Shell) Run(args []string) int {
	defer rescue()

	handleUsr1AndQuit()
	logSignals()

	if len(args) > 0 {
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "passing argument is not yet supported.")
			return 2
		}
		arg := args[0]
		if sh.cmd {
			sourceTextAndPrintError(sh.ev, "code from -c", arg)
		} else {
			script(sh.ev, arg)
		}
	} else if !sys.IsATTY(0) {
		script(sh.ev, "/dev/stdin")
	} else {
		interact(sh.ev, sh.daemon)
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
		util.PprintError(err)
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

func interact(ev *eval.Evaler, daemon *api.Client) {
	// Build Editor.
	sigch := make(chan os.Signal)
	signal.Notify(sigch)
	ed := edit.NewEditor(os.Stdin, os.Stderr, sigch, ev, daemon)

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
			logger.Println("signal", sig)
		}
	}()
}

func handleUsr1AndQuit() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGQUIT)
	go func() {
		for range sigs {
			fmt.Print(sys.DumpStack())
		}
	}()
}

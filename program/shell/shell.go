// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/elves/elvish/edit"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/runtime"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[shell] ")

// Shell keeps flags to the shell.
type Shell struct {
	BinPath     string
	SockPath    string
	DbPath      string
	Cmd         bool
	CompileOnly bool
}

func New(binpath, sockpath, dbpath string, cmd, compileonly bool) *Shell {
	return &Shell{binpath, sockpath, dbpath, cmd, compileonly}
}

// Main runs Elvish using the default terminal interface. It blocks until Elvish
// quites, and returns the exit code.
func (sh *Shell) Main(args []string) int {
	defer rescue()

	ev, dataDir := runtime.InitRuntime(sh.BinPath, sh.SockPath, sh.DbPath)
	defer runtime.CleanupRuntime(ev)

	handleSignals()

	if len(args) > 0 {
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "passing argument is not yet supported.")
			return 2
		}
		arg := args[0]
		if sh.CompileOnly {
			err := compileonly(ev, arg, sh.Cmd)
			if err != nil {
				util.PprintError(err)
				return 2
			}
		} else if sh.Cmd {
			err := ev.SourceText(eval.NewScriptSource("code from -c", "", arg))
			if err != nil {
				util.PprintError(err)
				return 2
			}
			return 0
		} else {
			return script(ev, arg)
		}
	} else {
		interact(ev, dataDir)
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

func script(ev *eval.Evaler, fname string) int {
	if !source(ev, fname, false) {
		return 1
	}
	return 0
}

func source(ev *eval.Evaler, fname string, notexistok bool) bool {
	abs, err := filepath.Abs(fname)
	var code string
	if err == nil {
		code, err = readFileUTF8(abs)
	}
	if err != nil {
		if notexistok && os.IsNotExist(err) {
			return true
		}
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	err = ev.SourceText(eval.NewScriptSource(fname, abs, code))
	if err != nil {
		util.PprintError(err)
		return false
	}
	return true
}

func compileonly(ev *eval.Evaler, arg string, command bool) error {
	var name, path, code string
	if command {
		name = "code from -c"
		path = ""
		code = arg
	} else {
		var err error
		name = arg
		path, err = filepath.Abs(name)
		if err != nil {
			return err
		}
		code, err = readFileUTF8(path)
		if err != nil {
			return err
		}
	}
	n, err := parse.Parse(name, code)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	_, err = ev.Compile(n, eval.NewScriptSource(name, path, code))
	return err
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

func interact(ev *eval.Evaler, dataDir string) {
	// Build Editor.
	var ed editor
	if sys.IsATTY(os.Stdin) {
		sigch := make(chan os.Signal)
		signal.Notify(sigch, syscall.SIGHUP, syscall.SIGINT, sys.SIGWINCH)
		ed = edit.NewEditor(os.Stdin, os.Stderr, sigch, ev)
	} else {
		ed = newMinEditor(os.Stdin, os.Stderr)
	}
	defer ed.Close()

	// Source rc.elv.
	if dataDir != "" {
		source(ev, dataDir+"/rc.elv", true)
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

		err = ev.SourceText(eval.NewInteractiveSource(line))
		if err != nil {
			util.PprintError(err)
		}
	}
}

func basicReadLine() (string, error) {
	stdin := bufio.NewReaderSize(os.Stdin, 0)
	return stdin.ReadString('\n')
}

func handleSignals() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	go func() {
		for sig := range sigs {
			logger.Println("signal", sig)
			handleSignal(sig)
		}
	}()
}

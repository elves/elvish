package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/elves/elvish/edit"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

func interact(ev *eval.Evaler, dataDir string, norc bool) {
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
	if !norc && dataDir != "" {
		err := sourceRC(ev, dataDir)
		if err != nil {
			util.PprintError(err)
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

		err = ev.EvalSource(eval.NewInteractiveSource(line))
		if err != nil {
			util.PprintError(err)
		}
	}
}

func sourceRC(ev *eval.Evaler, dataDir string) error {
	absPath, err := filepath.Abs(filepath.Join(dataDir, "rc.elv"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot get full path of rc.elv: %v", err)
	}
	code, err := readFileUTF8(absPath)

	return ev.SourceRC(eval.NewScriptSource("rc.elv", absPath, code))
}

func basicReadLine() (string, error) {
	stdin := bufio.NewReaderSize(os.Stdin, 0)
	return stdin.ReadString('\n')
}

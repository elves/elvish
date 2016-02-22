// Package stub is used to start and manage an elvish-stub process.
package stub

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/elves/elvish/util"
)

var stubname = "elvish-stub"

type Stub struct {
	process *os.Process
	// write is the other end of stdin of the stub.
	write *os.File
	// read is the other end of stdout of the stub.
	read    *os.File
	sigch   chan os.Signal
	statech chan struct{}
}

var stubEnv = []string{"A=BCDEFGHIJKLMNOPQRSTUVWXYZ"}

// NewStub spawns a new stub. The specified stderr is used for the subprocess.
func NewStub(stderr *os.File) (*Stub, error) {
	// Find stub.
	stubpath, err := searchStub()
	if err != nil {
		return nil, fmt.Errorf("search: %v", err)
	}

	// Make pipes.
	stdin, write, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("pipe: %v", err)
	}
	read, stdout, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("pipe: %v", err)
	}

	// Spawn stub.
	attr := os.ProcAttr{
		Env:   stubEnv,
		Files: []*os.File{stdin, stdout, stderr},
		Sys: &syscall.SysProcAttr{
			Setpgid: true,
		},
	}
	process, err := os.StartProcess(stubpath, []string{stubpath}, &attr)

	if err != nil {
		return nil, fmt.Errorf("spawn: %v", err)
	}

	// Wait for startup message.
	_, err = fmt.Fscanf(read, "ok\n")
	if err != nil {
		return nil, fmt.Errorf("read startup msg: %v", err)
	}

	// Spawn signal relayer and waiter.
	sigch := make(chan os.Signal)
	statech := make(chan struct{})
	go relaySignals(read, sigch)
	go wait(process, statech)

	return &Stub{process, write, read, sigch, statech}, nil
}

func searchStub() (string, error) {
	// os.Args[0] contains an absolute path. Find elvish-stub in the same
	// directory where elvish was started.
	if len(os.Args) > 0 && path.IsAbs(os.Args[0]) {
		stubpath := path.Join(path.Dir(os.Args[0]), stubname)
		if util.IsExecutable(stubpath) {
			return stubpath, nil
		}
	}
	return util.Search(strings.Split(os.Getenv("PATH"), ":"), stubname)
}

func (stub *Stub) Process() *os.Process {
	return stub.process
}

// Terminate terminates the stub.
func (stub *Stub) Terminate() {
	stub.write.Close()
}

// SetTitle sets the title of the stub.
func (stub *Stub) SetTitle(s string) {
	s = strings.TrimSpace(s)
	fmt.Fprintf(stub.write, "t %d %s\n", len(s), s)
}

func (stub *Stub) Chdir(dir string) {
	fmt.Fprintf(stub.write, "d %d %s\n", len(dir), dir)
}

// Signals returns a channel into which signals sent to the stub are relayed.
func (stub *Stub) Signals() <-chan os.Signal {
	return stub.sigch
}

// State returns a channel that is closed when the stub exits.
func (stub *Stub) State() <-chan struct{} {
	return stub.statech
}

// relaySignals relays output of the stub to sigch, assuming that outputs
// represent signal numbers.
func relaySignals(reader io.Reader, sigch chan<- os.Signal) {
	for {
		var signum int
		_, err := fmt.Fscanf(reader, "%d", &signum)
		if err != nil {
			sigch <- BadSignal{err}
			if err == io.EOF {
				break
			}
		} else {
			sigch <- syscall.Signal(signum)
		}
	}
}

func wait(proc *os.Process, ch chan<- struct{}) {
	for {
		state, err := proc.Wait()
		if err != nil || state.Exited() {
			break
		}
	}
	close(ch)
}

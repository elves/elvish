package shell

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/fsutil"
)

// Paths keeps all paths required for the Elvish runtime.
type Paths struct {
	RunDir string
	Sock   string

	DataDir string
	Db      string
	Rc      string
	LibDir  string

	Bin string
}

// MakePaths makes a populated Paths, using the given overrides.
func MakePaths(stderr io.Writer, overrides Paths) Paths {
	p := overrides
	setDir(&p.RunDir, "secure run directory", getSecureRunDir, stderr)
	if p.RunDir != "" {
		setChild(&p.Sock, p.RunDir, "sock")
	}

	setDir(&p.DataDir, "data directory", ensureDataDir, stderr)
	if p.DataDir != "" {
		setChild(&p.Db, p.DataDir, "db")
		setChild(&p.Rc, p.DataDir, "rc.elv")
		setChild(&p.LibDir, p.DataDir, "lib")
	}

	if p.Bin == "" {
		binFile, err := os.Executable()
		if err == nil {
			p.Bin = binFile
		} else {
			fmt.Fprintln(stderr, "warning: cannot get executable path:", err)
		}
	}
	return p
}

func setDir(p *string, what string, f func() (string, error), stderr io.Writer) {
	if *p == "" {
		dir, err := f()
		if err == nil {
			*p = dir
		} else {
			fmt.Fprintf(stderr, "warning: cannot create %v: %v\n", what, err)
		}
	}
}

func setChild(p *string, d, name string) {
	if *p == "" {
		*p = filepath.Join(d, name)
	}
}

// Ensures Elvish's data directory exists, creating it if necessary. It returns
// the path to the data directory (never with a trailing slash) and possible
// error.
func ensureDataDir() (string, error) {
	home, err := fsutil.GetHome("")
	if err != nil {
		return "", err
	}
	ddir := home + "/.elvish"
	return ddir, os.MkdirAll(ddir, 0700)
}

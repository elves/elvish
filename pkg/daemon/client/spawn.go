package client

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/fsutil"
)

// Spawn spawns a daemon process in the background by invoking BinPath, passing
// BinPath, DbPath and SockPath as command-line arguments after resolving them
// to absolute paths. The daemon log file is created in RunDir, and the stdout
// and stderr of the daemon is redirected to the log file.
//
// A suitable ProcAttr is chosen depending on the OS and makes sure that the
// daemon is detached from the current terminal, so that it is not affected by
// I/O or signals in the current terminal and keeps running after the current
// process quits.
func Spawn(cfg *daemondefs.SpawnConfig) error {
	binPath := cfg.BinPath
	// Determine binPath.
	if binPath == "" {
		bin, err := os.Executable()
		if err != nil {
			return errors.New("cannot find elvish: " + err.Error())
		}
		binPath = bin
	}

	var pathError error
	abs := func(name string, path string) string {
		if pathError != nil {
			return ""
		}
		if path == "" {
			pathError = fmt.Errorf("%s is required for spawning daemon", name)
			return ""
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			pathError = fmt.Errorf("cannot resolve %s to absolute path: %s", name, err)
		}
		return absPath
	}
	binPath = abs("BinPath", binPath)
	dbPath := abs("DbPath", cfg.DbPath)
	sockPath := abs("SockPath", cfg.SockPath)
	runDir := abs("RunDir", cfg.RunDir)
	if pathError != nil {
		return pathError
	}

	args := []string{
		binPath,
		"-daemon",
		"-bin", binPath,
		"-db", dbPath,
		"-sock", sockPath,
	}

	out, err := fsutil.ClaimFile(runDir, "daemon-*.log")
	if err != nil {
		return err
	}
	defer out.Close()

	// The daemon does not read any input; open DevNull and use it for stdin. We
	// could also just close the stdin, but on Unix that would make the first
	// file opened by the daemon take FD 0.
	in, err := os.OpenFile(os.DevNull, os.O_RDONLY, 0)
	if err != nil {
		in = os.Stdin
	} else {
		defer in.Close()
	}

	procattrs := procAttrForSpawn([]*os.File{in, out, out})
	_, err = os.StartProcess(binPath, args, procattrs)

	return err
}

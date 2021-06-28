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
	binPath, err := os.Executable()
	if err != nil {
		return errors.New("cannot find elvish: " + err.Error())
	}
	dbPath, err := abs("DbPath", cfg.DbPath)
	if err != nil {
		return err
	}
	sockPath, err := abs("SockPath", cfg.SockPath)
	if err != nil {
		return err
	}

	args := []string{
		binPath,
		"-daemon",
		"-db", dbPath,
		"-sock", sockPath,
	}

	// The daemon does not read any input; open DevNull and use it for stdin. We
	// could also just close the stdin, but on Unix that would make the first
	// file opened by the daemon take FD 0.
	in, err := os.OpenFile(os.DevNull, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := fsutil.ClaimFile(cfg.RunDir, "daemon-*.log")
	if err != nil {
		return err
	}
	defer out.Close()

	procattrs := procAttrForSpawn([]*os.File{in, out, out})

	_, err = os.StartProcess(binPath, args, procattrs)
	return err
}

func abs(name, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%s is required for spawning daemon", name)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve %s to absolute path: %s", name, err)
	}
	return absPath, nil
}

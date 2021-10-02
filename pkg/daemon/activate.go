package daemon

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/daemon/internal/api"
	"src.elv.sh/pkg/fsutil"
)

var (
	daemonSpawnTimeout     = time.Second
	daemonSpawnWaitPerLoop = 10 * time.Millisecond

	daemonKillTimeout     = time.Second
	daemonKillWaitPerLoop = 10 * time.Millisecond
)

type daemonStatus int

const (
	daemonOK daemonStatus = iota
	sockfileMissing
	sockfileOtherError
	connectionRefused
	connectionOtherError
	daemonOutdated
)

const connectionRefusedFmt = "Socket file %s exists but refuses requests. This is likely because the daemon was terminated abnormally. Going to remove socket file and re-spawn the daemon.\n"

// Activate returns a daemon client, either by connecting to an existing daemon,
// or spawning a new one. It always returns a non-nil client, even if there was an error.
func Activate(stderr io.Writer, spawnCfg *daemondefs.SpawnConfig) (daemondefs.Client, error) {
	sockpath := spawnCfg.SockPath
	cl := NewClient(sockpath)
	status, err := detectDaemon(sockpath, cl)
	shouldSpawn := false

	switch status {
	case daemonOK:
	case sockfileMissing:
		shouldSpawn = true
	case sockfileOtherError:
		return cl, fmt.Errorf("socket file %s inaccessible: %w", sockpath, err)
	case connectionRefused:
		fmt.Fprintf(stderr, connectionRefusedFmt, sockpath)
		err := os.Remove(sockpath)
		if err != nil {
			return cl, fmt.Errorf("failed to remove socket file: %w", err)
		}
		shouldSpawn = true
	case connectionOtherError:
		return cl, fmt.Errorf("unexpected RPC error on socket %s: %w", sockpath, err)
	case daemonOutdated:
		fmt.Fprintln(stderr, "Daemon is outdated; going to kill old daemon and re-spawn")
		err := killDaemon(sockpath, cl)
		if err != nil {
			return cl, fmt.Errorf("failed to kill old daemon: %w", err)
		}
		shouldSpawn = true
	default:
		return cl, fmt.Errorf("code bug: unknown daemon status %d", status)
	}

	if !shouldSpawn {
		return cl, nil
	}

	err = spawn(spawnCfg)
	if err != nil {
		return cl, fmt.Errorf("failed to spawn daemon: %w", err)
	}

	// Wait for daemon to come online
	start := time.Now()
	for time.Since(start) < daemonSpawnTimeout {
		cl.ResetConn()
		status, err := detectDaemon(sockpath, cl)

		switch status {
		case daemonOK:
			return cl, nil
		case sockfileMissing:
			// Continue waiting
		case sockfileOtherError:
			return cl, fmt.Errorf("socket file %s inaccessible: %w", sockpath, err)
		case connectionRefused:
			// Continue waiting
		case connectionOtherError:
			return cl, fmt.Errorf("unexpected RPC error on socket %s: %w", sockpath, err)
		case daemonOutdated:
			return cl, fmt.Errorf("code bug: newly spawned daemon is outdated")
		default:
			return cl, fmt.Errorf("code bug: unknown daemon status %d", status)
		}
		time.Sleep(daemonSpawnWaitPerLoop)
	}
	return cl, fmt.Errorf("daemon did not come up within %v", daemonSpawnTimeout)
}

func detectDaemon(sockpath string, cl daemondefs.Client) (daemonStatus, error) {
	_, err := os.Lstat(sockpath)
	if err != nil {
		if os.IsNotExist(err) {
			return sockfileMissing, err
		}
		return sockfileOtherError, err
	}

	version, err := cl.Version()
	if err != nil {
		if errors.Is(err, errConnRefused) {
			return connectionRefused, err
		}
		return connectionOtherError, err
	}
	if version < api.Version {
		return daemonOutdated, nil
	}
	return daemonOK, nil
}

func killDaemon(sockpath string, cl daemondefs.Client) error {
	pid, err := cl.Pid()
	if err != nil {
		return fmt.Errorf("kill daemon: %w", err)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("kill daemon: %w", err)
	}
	err = process.Signal(os.Interrupt)
	if err != nil {
		return fmt.Errorf("kill daemon: %w", err)
	}
	// Wait until the old daemon has removed the socket file, so that it doesn't
	// inadvertently remove the socket file of the new daemon we will start.
	start := time.Now()
	for time.Since(start) < daemonKillTimeout {
		_, err := os.Lstat(sockpath)
		if err == nil {
			time.Sleep(daemonKillWaitPerLoop)
		} else if os.IsNotExist(err) {
			return nil
		} else {
			return fmt.Errorf("kill daemon: %w", err)
		}
	}
	return fmt.Errorf("kill daemon: daemon did not remove socket within %v", daemonKillTimeout)
}

// Can be overridden in tests to avoid actual forking.
var startProcess = func(name string, argv []string, attr *os.ProcAttr) error {
	_, err := os.StartProcess(name, argv, attr)
	return err
}

// Spawns a daemon process in the background by invoking BinPath, passing
// BinPath, DbPath and SockPath as command-line arguments after resolving them
// to absolute paths. The daemon log file is created in RunDir, and the stdout
// and stderr of the daemon is redirected to the log file.
//
// A suitable ProcAttr is chosen depending on the OS and makes sure that the
// daemon is detached from the current terminal, so that it is not affected by
// I/O or signals in the current terminal and keeps running after the current
// process quits.
func spawn(cfg *daemondefs.SpawnConfig) error {
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

	err = startProcess(binPath, args, procattrs)
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

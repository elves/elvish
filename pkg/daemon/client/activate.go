package client

import (
	"errors"
	"fmt"
	"io"
	"net/rpc"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/daemon/internal/api"
)

const (
	daemonWaitLoops   = 100
	daemonWaitPerLoop = 10 * time.Millisecond
)

type daemonStatus int

const (
	daemonOK daemonStatus = iota
	sockfileMissing
	sockfileOtherError
	connectionShutdown
	connectionOtherError
	daemonInvalidDB
	daemonOutdated
)

const connectionShutdownFmt = "Socket file %s exists but is not responding to request. This is likely due to abnormal shutdown of the daemon. Going to remove socket file and re-spawn a daemon.\n"

var errInvalidDB = errors.New("daemon reported that database is invalid. If you upgraded Elvish from a pre-0.10 version, you need to upgrade your database by following instructions in https://github.com/elves/upgrade-db-for-0.10/")

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
		return cl, fmt.Errorf("socket file %s inaccessible: %v", sockpath, err)
	case connectionShutdown:
		fmt.Fprintf(stderr, connectionShutdownFmt, sockpath)
		err := os.Remove(sockpath)
		if err != nil {
			return cl, fmt.Errorf("failed to remove socket file: %v", err)
		}
		shouldSpawn = true
	case connectionOtherError:
		return cl, fmt.Errorf("unexpected RPC error on socket %s: %v", sockpath, err)
	case daemonInvalidDB:
		return cl, errInvalidDB
	case daemonOutdated:
		fmt.Fprintln(stderr, "Daemon is outdated; going to kill old daemon and re-spawn")
		err := killDaemon(cl)
		if err != nil {
			return cl, fmt.Errorf("failed to kill old daemon: %v", err)
		}
		shouldSpawn = true
	default:
		return cl, fmt.Errorf("code bug: unknown daemon status %d", status)
	}

	if !shouldSpawn {
		return cl, nil
	}

	err = Spawn(spawnCfg)
	if err != nil {
		return cl, fmt.Errorf("failed to spawn daemon: %v", err)
	}

	// Wait for daemon to come online
	for i := 0; i <= daemonWaitLoops; i++ {
		cl.ResetConn()
		status, err := detectDaemon(sockpath, cl)

		switch status {
		case daemonOK:
			return cl, nil
		case sockfileMissing:
			// Continue waiting
		case sockfileOtherError:
			return cl, fmt.Errorf("socket file %s inaccessible: %v", sockpath, err)
		case connectionShutdown:
			// Continue waiting
		case connectionOtherError:
			return cl, fmt.Errorf("unexpected RPC error on socket %s: %v", sockpath, err)
		case daemonInvalidDB:
			return cl, errInvalidDB
		case daemonOutdated:
			return cl, fmt.Errorf("code bug: newly spawned daemon is outdated")
		default:
			return cl, fmt.Errorf("code bug: unknown daemon status %d", status)
		}
		time.Sleep(daemonWaitPerLoop)
	}
	return cl, fmt.Errorf("daemon unreachable after waiting for %s", daemonWaitLoops*daemonWaitPerLoop)
}

func detectDaemon(sockpath string, cl daemondefs.Client) (daemonStatus, error) {
	_, err := os.Stat(sockpath)
	if err != nil {
		if os.IsNotExist(err) {
			return sockfileMissing, err
		}
		return sockfileOtherError, err
	}

	version, err := cl.Version()
	if err != nil {
		switch {
		case err == rpc.ErrShutdown:
			return connectionShutdown, err
		case err.Error() == bolt.ErrInvalid.Error():
			return daemonInvalidDB, err
		default:
			return connectionOtherError, err
		}
	}
	if version < api.Version {
		return daemonOutdated, nil
	}
	return daemonOK, nil
}

func killDaemon(cl daemondefs.Client) error {
	pid, err := cl.Pid()
	if err != nil {
		return fmt.Errorf("cannot get pid of daemon: %v", err)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("cannot find daemon process (pid=%d): %v", pid, err)
	}
	return process.Signal(os.Interrupt)
}

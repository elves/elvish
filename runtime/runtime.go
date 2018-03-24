// Package runtime assembles the Elvish runtime.
package runtime

import (
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval"
	daemonmod "github.com/elves/elvish/eval/daemon"
	"github.com/elves/elvish/eval/re"
	storemod "github.com/elves/elvish/eval/store"
	"github.com/elves/elvish/eval/str"
	daemonp "github.com/elves/elvish/program/daemon"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[runtime] ")

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

const (
	daemonWontWorkMsg     = "Daemon-related functions will likely not work."
	connectionShutdownFmt = "Socket file %s exists but is not responding to request. This is likely due to abnormal shutdown of the daemon. Going to remove socket file and re-spawn a daemon.\n"
)

var errInvalidDB = errors.New("daemon reported that database is invalid. If you upgraded Elvish from a pre-0.10 version, you need to upgrade your database by following instructions in https://github.com/elves/upgrade-db-for-0.10/")

// InitRuntime initializes the runtime. The caller is responsible for calling
// CleanupRuntime at some point.
func InitRuntime(binpath, sockpath, dbpath string) (*eval.Evaler, string) {
	var dataDir string
	var err error

	// Determine data directory.
	dataDir, err = storedefs.EnsureDataDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: cannot create data directory ~/.elvish")
	} else {
		if dbpath == "" {
			dbpath = filepath.Join(dataDir, "db")
		}
	}

	// Determine runtime directory.
	runDir, err := getSecureRunDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot get runtime dir /tmp/elvish-$uid, falling back to data dir ~/.elvish:", err)
		runDir = dataDir
	}
	if sockpath == "" {
		sockpath = filepath.Join(runDir, "sock")
	}

	ev := eval.NewEvaler()
	ev.SetLibDir(filepath.Join(dataDir, "lib"))
	ev.InstallModule("re", re.Ns)
	ev.InstallModule("str", str.Ns)
	if sockpath != "" && dbpath != "" {
		spawner := &daemonp.Daemon{
			BinPath:       binpath,
			DbPath:        dbpath,
			SockPath:      sockpath,
			LogPathPrefix: filepath.Join(runDir, "daemon.log-"),
		}
		// TODO(xiaq): Connect to daemon and install daemon module
		// asynchronously.
		client, err := connectToDaemon(sockpath, spawner)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot connect to daemon:", err)
			fmt.Fprintln(os.Stderr, daemonWontWorkMsg)
		}
		// Even if error is not nil, we install daemon-related functionalities
		// anyway. Daemon may eventually come online and become functional.
		ev.InstallDaemonClient(client)
		ev.InstallModule("store", storemod.Ns(client))
		ev.InstallModule("daemon", daemonmod.Ns(client, spawner))
	}
	return ev, dataDir
}

func connectToDaemon(sockpath string, spawner *daemonp.Daemon) (*daemon.Client, error) {
	cl := daemon.NewClient(sockpath)
	status, err := detectDaemon(sockpath, cl)
	shouldSpawn := false

	switch status {
	case daemonOK:
	case sockfileMissing:
		shouldSpawn = true
	case sockfileOtherError:
		return cl, fmt.Errorf("socket file %s inaccessible: %v", sockpath, err)
	case connectionShutdown:
		fmt.Fprintf(os.Stderr, connectionShutdownFmt, sockpath)
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
		fmt.Fprintln(os.Stderr, "Daemon is outdated; going to kill old daemon and re-spawn")
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

	err = spawner.Spawn()
	if err != nil {
		return cl, fmt.Errorf("failed to spawn daemon: %v", err)
	}
	logger.Println("Spawned daemon")

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

func detectDaemon(sockpath string, cl *daemon.Client) (daemonStatus, error) {
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
	if version < daemon.Version {
		return daemonOutdated, nil
	}
	return daemonOK, nil
}

func killDaemon(cl *daemon.Client) error {
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

// CleanupRuntime cleans up the runtime.
func CleanupRuntime(ev *eval.Evaler) {
	if ev.DaemonClient != nil {
		err := ev.DaemonClient.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: failed to close connection to daemon:", err)
		}
	}
	ev.Close()
}

var (
	ErrBadOwner      = errors.New("bad owner")
	ErrBadPermission = errors.New("bad permission")
)

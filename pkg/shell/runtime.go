package shell

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/eval"
	daemonmod "src.elv.sh/pkg/eval/mods/daemon"
	mathmod "src.elv.sh/pkg/eval/mods/math"
	pathmod "src.elv.sh/pkg/eval/mods/path"
	"src.elv.sh/pkg/eval/mods/platform"
	"src.elv.sh/pkg/eval/mods/re"
	"src.elv.sh/pkg/eval/mods/store"
	"src.elv.sh/pkg/eval/mods/str"
	"src.elv.sh/pkg/eval/mods/unix"
	"src.elv.sh/pkg/rpc"
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

const (
	daemonWontWorkMsg     = "Daemon-related functions will likely not work."
	connectionShutdownFmt = "Socket file %s exists but is not responding to request. This is likely due to abnormal shutdown of the daemon. Going to remove socket file and re-spawn a daemon.\n"
)

var errInvalidDB = errors.New("daemon reported that database is invalid. If you upgraded Elvish from a pre-0.10 version, you need to upgrade your database by following instructions in https://github.com/elves/upgrade-db-for-0.10/")

// InitRuntime initializes the runtime. The caller should call CleanupRuntime
// when the Evaler is no longer needed.
func InitRuntime(stderr io.Writer, p Paths, spawn bool) *eval.Evaler {
	ev := eval.NewEvaler()
	ev.SetLibDir(p.LibDir)
	ev.AddModule("math", mathmod.Ns)
	ev.AddModule("path", pathmod.Ns)
	ev.AddModule("platform", platform.Ns)
	ev.AddModule("re", re.Ns)
	ev.AddModule("str", str.Ns)
	if unix.ExposeUnixNs {
		ev.AddModule("unix", unix.Ns)
	}

	if spawn && p.Sock != "" && p.Db != "" {
		spawnCfg := &daemon.SpawnConfig{
			RunDir:   p.RunDir,
			BinPath:  p.Bin,
			DbPath:   p.Db,
			SockPath: p.Sock,
		}
		// TODO(xiaq): Connect to daemon and install daemon module
		// asynchronously.
		client, err := connectToDaemon(stderr, spawnCfg)
		if err != nil {
			fmt.Fprintln(stderr, "Cannot connect to daemon:", err)
			fmt.Fprintln(stderr, daemonWontWorkMsg)
		}
		// Even if error is not nil, we install daemon-related functionalities
		// anyway. Daemon may eventually come online and become functional.
		ev.SetDaemonClient(client)
		ev.AddModule("store", store.Ns(client))
		ev.AddModule("daemon", daemonmod.Ns(client, spawnCfg))
	}
	return ev
}

// CleanupRuntime cleans up the runtime.
func CleanupRuntime(stderr io.Writer, ev *eval.Evaler) {
	daemon := ev.DaemonClient()
	if daemon != nil {
		err := daemon.Close()
		if err != nil {
			fmt.Fprintln(stderr,
				"warning: failed to close connection to daemon:", err)
		}
	}
}

func connectToDaemon(stderr io.Writer, spawnCfg *daemon.SpawnConfig) (daemon.Client, error) {
	sockpath := spawnCfg.SockPath
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

	err = daemon.Spawn(spawnCfg)
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

func detectDaemon(sockpath string, cl daemon.Client) (daemonStatus, error) {
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

func killDaemon(cl daemon.Client) error {
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

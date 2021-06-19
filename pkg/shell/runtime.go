package shell

import (
	"fmt"
	"io"

	"src.elv.sh/pkg/daemon/client"
	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/eval"
	daemonmod "src.elv.sh/pkg/eval/mods/daemon"
	"src.elv.sh/pkg/eval/mods/file"
	mathmod "src.elv.sh/pkg/eval/mods/math"
	pathmod "src.elv.sh/pkg/eval/mods/path"
	"src.elv.sh/pkg/eval/mods/platform"
	"src.elv.sh/pkg/eval/mods/re"
	"src.elv.sh/pkg/eval/mods/store"
	"src.elv.sh/pkg/eval/mods/str"
	"src.elv.sh/pkg/eval/mods/unix"
)

const (
	daemonWontWorkMsg = "Daemon-related functions will likely not work."
)

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
	ev.AddModule("file", file.Ns)
	if unix.ExposeUnixNs {
		ev.AddModule("unix", unix.Ns)
	}

	if spawn && p.Sock != "" && p.Db != "" {
		spawnCfg := &daemondefs.SpawnConfig{
			RunDir:   p.RunDir,
			BinPath:  p.Bin,
			DbPath:   p.Db,
			SockPath: p.Sock,
		}
		// TODO(xiaq): Connect to daemon and install daemon module
		// asynchronously.
		cl, err := client.Activate(stderr, spawnCfg)
		if err != nil {
			fmt.Fprintln(stderr, "Cannot connect to daemon:", err)
			fmt.Fprintln(stderr, daemonWontWorkMsg)
		}
		// Even if error is not nil, we install daemon-related functionalities
		// anyway. Daemon may eventually come online and become functional.
		ev.SetDaemonClient(cl)
		ev.AddModule("store", store.Ns(cl))
		ev.AddModule("daemon", daemonmod.Ns(cl))
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

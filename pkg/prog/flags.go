package prog

import "flag"

// FlagSet wraps a [flag.FlagSet], and provides methods to register flags shared
// by multiple subprograms on demand.
type FlagSet struct {
	*flag.FlagSet
	daemonPaths *DaemonPaths
	json        *bool
}

// DaemonPaths stores the -db and -sock flags.
type DaemonPaths struct {
	DB, Sock string
}

// DaemonPaths returns a pointer to a struct storing the value of -db and
// -sock flags, registering them on demand.
func (fs *FlagSet) DaemonPaths() *DaemonPaths {
	if fs.daemonPaths == nil {
		var dp DaemonPaths
		fs.StringVar(&dp.DB, "db", "",
			"[internal flag] Path to the database file")
		fs.StringVar(&dp.Sock, "sock", "",
			"[internal flag] Path to the daemon's Unix socket")
		fs.daemonPaths = &dp
	}
	return fs.daemonPaths
}

// JSON returns a pointer to the value of the -json flag, registering it on
// demand.
func (fs *FlagSet) JSON() *bool {
	if fs.json == nil {
		fs.json = fs.Bool("json", false,
			"Show the output from -buildinfo, -compileonly or -version in JSON")
	}
	return fs.json
}

package prog

import "flag"

type FlagSet struct {
	*flag.FlagSet
	daemonPaths *DaemonPaths
	json        *bool
}

type DaemonPaths struct {
	DB, Sock string
}

func (fs *FlagSet) DaemonPaths() *DaemonPaths {
	if fs.daemonPaths == nil {
		var dp DaemonPaths
		fs.StringVar(&dp.DB, "db", "",
			"[internal flag] Path to the database file")
		fs.StringVar(&dp.Sock, "sock", "",
			"[internal flag] Path to the daemon's UNIX socket")
		fs.daemonPaths = &dp
	}
	return fs.daemonPaths
}

func (fs *FlagSet) JSON() *bool {
	if fs.json == nil {
		var json bool
		fs.BoolVar(&json, "json", false,
			"Show the output from -buildinfo, -compileonly or -version in JSON")
		fs.json = &json
	}
	return fs.json
}

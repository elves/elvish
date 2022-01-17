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
		fs.StringVar(&dp.DB, "db", "", "[internal flag] path to the database")
		fs.StringVar(&dp.Sock, "sock", "", "[internal flag] path to the daemon socket")
		fs.daemonPaths = &dp
	}
	return fs.daemonPaths
}

func (fs *FlagSet) JSON() *bool {
	if fs.json == nil {
		var json bool
		fs.BoolVar(&json, "json", false, "show output in JSON. Useful with -buildinfo, -version and -compileonly")
		fs.json = &json
	}
	return fs.json
}

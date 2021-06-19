// Package daemondefs contains definitions used for the daemon.
//
// It is a separate package so that packages that only depend on the daemon
// API does not need to depend on the concrete implementation.
package daemondefs

import (
	"io"

	"src.elv.sh/pkg/store"
)

// Client represents a daemon client.
type Client interface {
	store.Store

	ResetConn() error
	Close() error

	Pid() (int, error)
	SockPath() string
	Version() (int, error)
}

// ActivateFunc is a function that activates a daemon client, possibly by
// spawning a new daemon and connecting to it.
type ActivateFunc func(stderr io.Writer, spawnCfg *SpawnConfig) (Client, error)

// SpawnConfig keeps configurations for spawning the daemon.
type SpawnConfig struct {
	// BinPath is the path to the Elvish binary itself, used when forking. This
	// field is used only when spawning the daemon. If empty, it is
	// automatically determined with os.Executable.
	BinPath string
	// DbPath is the path to the database.
	DbPath string
	// SockPath is the path to the socket on which the daemon will serve
	// requests.
	SockPath string
	// RunDir is the directory in which to place the daemon log file.
	RunDir string
}

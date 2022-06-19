package fsutil

import (
	"os"
	"path/filepath"

	"src.elv.sh/pkg/env"
)

// ConfigHome returns the directory that is searched for the Elvish `rc.elv` script. It corresponds
// to the XDG_CONFIG_HOME env var, but includes the `elvish` subdir, whether or not the env var is
// explicitly set.
func ConfigHome() (string, error) {
	if configHome := os.Getenv(env.XDG_CONFIG_HOME); configHome != "" {
		return filepath.Join(configHome, "elvish"), nil
	}
	return DefaultConfigHome()
}

// DataHome returns the directory that is searched for the Elvish `rc.elv` script. It corresponds
// to the XDG_DATA_HOME env var, but includes the `elvish` subdir, whether or not the env var is
// explicitly set.
func DataHome() (string, error) {
	if dataHome := os.Getenv(env.XDG_DATA_HOME); dataHome != "" {
		return filepath.Join(dataHome, "elvish"), nil
	}
	return DefaultDataHome()
}

// StateHome returns the directory that is used for the Elvish database. It corresponds to the
// XDG_STATE_HOME env var, but includes the `elvish` subdir, whether or not the env var is
// explicitly set.
func StateHome() (string, error) {
	if stateHome := os.Getenv(env.XDG_STATE_HOME); stateHome != "" {
		return filepath.Join(stateHome, "elvish"), nil
	}
	return DefaultStateHome()
}

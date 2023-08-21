// Package env keeps names of environment variables with special significance to
// Elvish.
package env

// Environment variables with special significance to Elvish.
const (
	HOME      = "HOME"
	LS_COLORS = "LS_COLORS"
	NO_COLOR  = "NO_COLOR"
	PATH      = "PATH"
	PWD       = "PWD"
	SHLVL     = "SHLVL"
	USERNAME  = "USERNAME"

	// Only used on Unix
	XDG_CONFIG_HOME = "XDG_CONFIG_HOME"
	XDG_DATA_DIRS   = "XDG_DATA_DIRS"
	XDG_DATA_HOME   = "XDG_DATA_HOME"
	XDG_RUNTIME_DIR = "XDG_RUNTIME_DIR"
	XDG_STATE_HOME  = "XDG_STATE_HOME"

	// Only used on Windows
	PATHEXT = "PATHEXT"

	// Only used in tests
	ELVISH_TEST_TIME_SCALE = "ELVISH_TEST_TIME_SCALE"
)

// Package env keeps names of environment variables with special significance to
// Elvish.
package env

// Environment variables with special significance to Elvish.
//
// Note that some of these env vars may be significant only in special
// circumstances, such as when running unit tests.
const (
	ELVISH_TEST_TIME_SCALE = "ELVISH_TEST_TIME_SCALE"
	HOME                   = "HOME"
	LS_COLORS              = "LS_COLORS"
	PATH                   = "PATH"
	PATHEXT                = "PATHEXT"
	PWD                    = "PWD"
	SHLVL                  = "SHLVL"
	USERNAME               = "USERNAME"
	XDG_RUNTIME_DIR        = "XDG_RUNTIME_DIR"
)

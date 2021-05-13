// Package env keeps names of environment variables with special significance to
// Elvish.
package env

// Environment variables with special significance to Elvish.
//
// Note that some of these env vars may be significant only in special
// circumstances, such as when running unit tests.
const (
	ElvishTestTimeScale = "ELVISH_TEST_TIME_SCALE"
	Home                = "HOME"
	LsColors            = "LS_COLORS"
	Path                = "PATH"
	Pathext             = "PATHEXT"
	Pwd                 = "PWD"
	Shlvl               = "SHLVL"
	Username            = "USERNAME"
	XdgRuntimeDir       = "XDG_RUNTIME_DIR"
)

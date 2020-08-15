// Package util contains utility functions and constants.
package util

// Environment variables with special significance to Elvish.
//
// Note that some of these env vars may be significant only in special
// circumstances; such as when running unit tests.
const (
	EnvELVISH_TEST_TIME_SCALE = "ELVISH_TEST_TIME_SCALE"
	EnvHOME                   = "HOME"
	EnvLS_COLORS              = "LS_COLORS"
	EnvPATH                   = "PATH"
	EnvPATHEXT                = "PATHEXT"
	EnvPWD                    = "PWD"
	EnvSHLVL                  = "SHLVL"
	EnvUSERNAME               = "USERNAME"
	EnvXDG_RUNTIME_DIR        = "XDG_RUNTIME_DIR"
)

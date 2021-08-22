//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package platform

const (
	isUnix    = true
	isWindows = false
)

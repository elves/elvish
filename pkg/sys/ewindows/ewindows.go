//go:generate cmd /c go tool cgo -godefs types.go > ztypes_windows.go && gofmt -w ztypes_windows.go
//go:build windows

// Package ewindows provides extra Windows-specific system utilities.
package ewindows

import "golang.org/x/sys/windows"

var kernel32 = windows.NewLazySystemDLL("kernel32.dll")

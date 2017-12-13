//go:generate cmd /c go tool cgo -godefs types_src_windows.go > ztypes_windows.go && gofmt -w ztypes_windows.go

package sys

//go:build darwin && (amd64 || arm64)
// +build darwin
// +build amd64 arm64

package eunix

// Only Darwin uses 64-bit flags in Termios on 64-bit architectures. See:
// https://cs.opensource.google/search?q=%5BIOCL%5Dflag.*uint64&sq=&ss=go%2Fx%2Fsys

type termiosFlag = uint64

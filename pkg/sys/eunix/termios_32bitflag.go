//go:build !(darwin && (amd64 || arm64))
// +build !darwin !amd64,!arm64

package eunix

type termiosFlag = uint32

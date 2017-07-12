// +build !cgo

// Package service implements the daemon service for mediating access to the
// storage backend.
package service

import (
	"github.com/elves/elvish/util"
	"syscall"
)

var logger = util.GetLogger("[daemon-dummy] ")

// A dummy implementation of Serve for the environment of Cgo being disabled.
func Serve(sockpath, dbpath string) {
	logger.Println("pid is", syscall.Getpid())
	logger.Println("going to listen", sockpath)

	select {}

	logger.Println("exiting")
}

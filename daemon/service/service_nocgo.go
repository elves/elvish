// +build !cgo

// Package service implements the daemon service for mediating access to the
// storage backend.
package service

import (
	"syscall"

	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[daemon-dummy] ")

// A dummy implementation of Serve for the environment of Cgo being disabled.
func Serve(sockpath, dbpath string) {
	logger.Println("pid is", syscall.Getpid())
	logger.Println("this is the dummy service implementation", sockpath)

	select {}

	logger.Println("exiting")
}

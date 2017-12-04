// Package build contains build information. They can be set during compilation
// by passing -ldflags "-X github.com/elves/elvish/build.Var=value"
// to "go build" or "go get".

package build

var Version = "unknown"

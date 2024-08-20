// Package storedefs contains definitions of the store API.
//
// It is a separate package so that packages that only depend on the store API
// does not need to depend on the concrete implementation.
package storedefs

import "errors"

// NoBlacklist is an empty blacklist, to be used in GetDirs.
var NoBlacklist = map[string]struct{}{}

// ErrNoMatchingCmd is the error returned when a LastCmd or FirstCmd query
// completes with no result.
var ErrNoMatchingCmd = errors.New("no matching command line")

// Store is an interface satisfied by the storage service.
type Store interface {
	NextCmdSeq() (int, error)
	AddCmd(text string) (int, error)
	DelCmd(seq int) error
	Cmd(seq int) (string, error)
	CmdsWithSeq(from, upto int) ([]Cmd, error)
	NextCmd(from int, prefix string) (Cmd, error)
	PrevCmd(upto int, prefix string) (Cmd, error)

	AddDir(dir string, incFactor float64) error
	DelDir(dir string) error
	Dirs(blacklist map[string]struct{}) ([]Dir, error)
}

// Dir is an entry in the directory history.
type Dir struct {
	Path  string
	Score float64
}

// Cmd is an entry in the command history.
type Cmd struct {
	Text string
	Seq  int
}

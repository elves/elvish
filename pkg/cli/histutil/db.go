package histutil

import (
	"src.elv.sh/pkg/store/storedefs"
)

// DB is the interface of the storage database.
type DB interface {
	NextCmdSeq() (int, error)
	AddCmd(cmd string) (int, error)
	CmdsWithSeq(from, upto int) ([]storedefs.Cmd, error)
	PrevCmd(upto int, prefix string) (storedefs.Cmd, error)
	NextCmd(from int, prefix string) (storedefs.Cmd, error)
}

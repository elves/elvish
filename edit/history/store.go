package history

// Store is the interface of the storage backend.
type Store interface {
	NextCmdSeq() (int, error)
	AddCmd(cmd string) (int, error)
	Cmds(from, upto int) ([]string, error)
	PrevCmd(upto int, prefix string) (int, string, error)
}

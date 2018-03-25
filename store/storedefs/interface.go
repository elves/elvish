package storedefs

// Store is an interface satisfied by the storage service.
type Store interface {
	NextCmdSeq() (int, error)
	AddCmd(text string) (int, error)
	DelCmd(seq int) error
	Cmd(seq int) (string, error)
	Cmds(from, upto int) ([]string, error)
	NextCmd(from int, prefix string) (int, string, error)
	PrevCmd(upto int, prefix string) (int, string, error)

	AddDir(dir string, incFactor float64) error
	DelDir(dir string) error
	Dirs(blacklist map[string]struct{}) ([]Dir, error)

	SharedVar(name string) (string, error)
	SetSharedVar(name, value string) error
	DelSharedVar(name string) error
}

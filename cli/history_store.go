package cli

// HistoryStore represents a store for command history.
type HistoryStore interface {
	AddCmd(cmd string) error
	AllCmds() ([]string, error)
}

// MemoryHistoryStore returns a HistoryStore that stores command history in
// memory.
func MemoryHistoryStore() HistoryStore {
	return &memoryHistoryStore{}
}

type memoryHistoryStore struct {
	cmds []string
}

func (hs *memoryHistoryStore) AddCmd(cmd string) error {
	hs.cmds = append(hs.cmds, cmd)
	return nil
}

func (hs *memoryHistoryStore) AllCmds() ([]string, error) {
	return hs.cmds, nil
}

package parse

// Redir represents a single IO redirection. Its concrete type may be one of
// the *Redir types below.
type Redir interface {
	Node
	Fd() uintptr
	// ensure only structs in this package can satisfy this interface
	unexported()
}

type RedirBase struct {
	Pos
	FD uintptr
}

func (r *RedirBase) Fd() uintptr {
	return r.FD
}

func (r *RedirBase) unexported() {
}

// FdRedir represents redirection into another fd, like >[2=3].
type FdRedir struct {
	RedirBase
	OldFd uintptr
}

// NewFdRedir creates a new FdRedir. Public since we need to turn FilenameRedir
// -> FdRedir when evaluating commands.
func NewFdRedir(pos Pos, fd, oldFd uintptr) *FdRedir {
	return &FdRedir{RedirBase{pos, fd}, oldFd}
}

func (fr *FdRedir) isNode() {}

// CloseRedir represents the closing of a fd, like >[2=].
type CloseRedir struct {
	RedirBase
}

func newCloseRedir(pos Pos, fd uintptr) *CloseRedir {
	return &CloseRedir{RedirBase{pos, fd}}
}

func (cr *CloseRedir) isNode() {}

// FilenameRedir represents redirection into a file, like >a.txt
type FilenameRedir struct {
	RedirBase
	Flag     int
	Filename *Compound
}

func newFilenameRedir(pos Pos, fd uintptr, flag int, filename *Compound) *FilenameRedir {
	return &FilenameRedir{RedirBase{pos, fd}, flag, filename}
}

func (fr *FilenameRedir) isNode() {}

package parse

// Redir represents a single IO redirection. Its concrete type may be one of
// the *Redir types below.
type Redir interface {
	Node
	Fd() uintptr
	// ensure only structs in this package can satisfy this interface
	isRedir()
}

// redirBase is the common part of all Redir structs. It implements the Redir
// interface.
type redirBase struct {
	Pos
	FD uintptr // the fd to operate on.
}

func (r *redirBase) Fd() uintptr {
	return r.FD
}

func (r *redirBase) isRedir() {}

func (r *redirBase) isNode() {}

// FdRedir represents redirection into another fd, like >[2=3].
type FdRedir struct {
	redirBase
	OldFd uintptr
}

// NewFdRedir creates a new FdRedir. Public since we need to turn FilenameRedir
// -> FdRedir when evaluating commands.
func NewFdRedir(pos Pos, fd, oldFd uintptr) *FdRedir {
	return &FdRedir{redirBase{pos, fd}, oldFd}
}

// CloseRedir represents the closing of a fd, like >[2=].
type CloseRedir struct {
	redirBase
}

func newCloseRedir(pos Pos, fd uintptr) *CloseRedir {
	return &CloseRedir{redirBase{pos, fd}}
}

// FilenameRedir represents redirection into a file, like >a.txt
type FilenameRedir struct {
	redirBase
	Flag     int
	Filename *Compound
}

func newFilenameRedir(pos Pos, fd uintptr, flag int, filename *Compound) *FilenameRedir {
	return &FilenameRedir{redirBase{pos, fd}, flag, filename}
}

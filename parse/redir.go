package parse

// Redir represents a single IO redirection. Its concrete type may be one of
// the *Redir types below.
type Redir interface {
	Node
	Fd() uintptr
	// ensure only structs in this package can satisfy this interface
	unexported()
}

type redir struct {
	Pos
	fd uintptr
}

func (r *redir) Fd() uintptr {
	return r.fd
}

func (r *redir) unexported() {
}

type FdRedir struct {
	redir
	OldFd uintptr
}

// Public since we need to turn FilenameRedir -> FdRedir when evaluating
// commands.
func NewFdRedir(pos Pos, fd, oldFd uintptr) *FdRedir {
	return &FdRedir{redir{pos, fd}, oldFd}
}

type CloseRedir struct {
	redir
}

func newCloseRedir(pos Pos, fd uintptr) *CloseRedir {
	return &CloseRedir{redir{pos, fd}}
}

type FilenameRedir struct {
	redir
	Flag int
	Filename *ListNode // a Term
}

func newFilenameRedir(pos Pos, fd uintptr, flag int, filename *ListNode) *FilenameRedir {
	return &FilenameRedir{redir{pos, fd}, flag, filename}
}

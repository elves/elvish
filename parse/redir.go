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

// FdRedir represents redirection into another fd, like >[2=3].
type FdRedir struct {
	redir
	OldFd uintptr
}

// NewFdRedir creates a new FdRedir. Public since we need to turn FilenameRedir
// -> FdRedir when evaluating commands.
func NewFdRedir(pos Pos, fd, oldFd uintptr) *FdRedir {
	return &FdRedir{redir{pos, fd}, oldFd}
}

func (fr *FdRedir) isNode() {}

// CloseRedir represents the closing of a fd, like >[2=].
type CloseRedir struct {
	redir
}

func newCloseRedir(pos Pos, fd uintptr) *CloseRedir {
	return &CloseRedir{redir{pos, fd}}
}

func (cr *CloseRedir) isNode() {}

// FilenameRedir represents redirection into a file, like >a.txt
type FilenameRedir struct {
	redir
	Flag     int
	Filename *ListNode // a Term
}

func newFilenameRedir(pos Pos, fd uintptr, flag int, filename *ListNode) *FilenameRedir {
	return &FilenameRedir{redir{pos, fd}, flag, filename}
}

func (fr *FilenameRedir) isNode() {}

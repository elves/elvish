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

func (fr *FdRedir) Isomorph(n Node) bool {
	if fr2, ok := n.(*FdRedir); ok {
		return fr.fd == fr2.fd && fr.OldFd == fr2.OldFd
	}
	return false
}

type CloseRedir struct {
	redir
}

func newCloseRedir(pos Pos, fd uintptr) *CloseRedir {
	return &CloseRedir{redir{pos, fd}}
}

func (cr *CloseRedir) Isomorph(n Node) bool {
	if cr2, ok := n.(*CloseRedir); ok {
		return cr.fd == cr2.fd
	}
	return false
}

type FilenameRedir struct {
	redir
	Flag     int
	Filename *ListNode // a Term
}

func newFilenameRedir(pos Pos, fd uintptr, flag int, filename *ListNode) *FilenameRedir {
	return &FilenameRedir{redir{pos, fd}, flag, filename}
}

func (fr *FilenameRedir) Isomorph(n Node) bool {
	if fr2, ok := n.(*FilenameRedir); ok {
		return fr.fd == fr2.fd && fr.Flag == fr2.Flag && fr.Filename.Isomorph(fr2.Filename)
	}
	return false
}

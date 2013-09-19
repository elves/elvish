package parse

// Redir represents a single IO redirection. Its concrete type may be one of
// the *Redir types below.
type Redir interface {
	Fd() int
	// ensure only structs in this package can satisfy this interface
	unexported()
}

type redir struct {
	fd int
}

func (r *redir) Fd() int {
	return r.fd
}

func (r *redir) unexported() {
}

type FdRedir struct {
	redir
	OldFd int
}

// Public since we need to turn FilenameRedir -> FdRedir when evaluating
// commands.
func NewFdRedir(fd int, oldFd int) *FdRedir {
	return &FdRedir{redir{fd}, oldFd}
}

type CloseRedir struct {
	redir
}

func newCloseRedir(fd int) *CloseRedir {
	return &CloseRedir{redir{fd}}
}

type FilenameRedir struct {
	redir
	Flag int
	Filename Node
}

func newFilenameRedir(fd int, flag int, filename Node) *FilenameRedir {
	return &FilenameRedir{redir{fd}, flag, filename}
}

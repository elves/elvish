package parse

// Redir represents a single IO redirection. Its concrete type may be one of
// the *Redir types below.
type Redir interface {
	Oldfd() int
	unexported()
}

type redir struct {
	oldfd int
}

func (r *redir) Oldfd() int {
	return r.oldfd
}

func (r *redir) unexported() {
}

type FdRedir struct {
	redir
	Newfd int
}

func newFdRedir(oldfd int, newfd int) *FdRedir {
	return &FdRedir{redir{oldfd}, newfd}
}

type CloseRedir struct {
	redir
}

func newCloseRedir(oldfd int) *CloseRedir {
	return &CloseRedir{redir{oldfd}}
}

type FilenameRedir struct {
	redir
	Flag int
	Filename Node
}

func newFilenameRedir(oldfd int, flag int, filename Node) *FilenameRedir {
	return &FilenameRedir{redir{oldfd}, flag, filename}
}

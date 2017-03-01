package api

// Basic requests.

type GetPid struct{}

// Cmd requests.

type NextCmdSeq struct{}

type AddCmd struct {
	Text string
}

type GetCmds struct {
	From int
	Upto int
}

type GetFirstCmd struct {
	From   int
	Prefix string
}

type GetLastCmd struct {
	Upto   int
	Prefix string
}

// Dir requests.

type AddDir struct {
	Dir       string
	IncFactor float64
}

type GetDirs struct {
	Blacklist map[string]struct{}
}

// SharedVar requests.

type GetSharedVar struct {
	Name string
}

type SetSharedVar struct {
	Name  string
	Value string
}

type DelSharedVar struct {
	Name string
}

type Request struct {
	GetPid *GetPid

	NextCmdSeq  *NextCmdSeq
	AddCmd      *AddCmd
	GetCmds     *GetCmds
	GetFirstCmd *GetFirstCmd
	GetLastCmd  *GetLastCmd

	AddDir  *AddDir
	GetDirs *GetDirs

	GetSharedVar *GetSharedVar
	SetSharedVar *SetSharedVar
	DelSharedVar *DelSharedVar
}

type ResponseHeader struct {
	Error   *string `json:",omitempty"`
	Sending *int    `json:",omitempty"`
}

func (h *ResponseHeader) OK() bool {
	return h.Error == nil
}

package api

type GetPid struct{}

type AddDir struct {
	Dir       string
	IncFactor float64
}

type ListDirs struct {
	Blacklist map[string]struct{}
}

type Request struct {
	GetPid   *GetPid
	AddDir   *AddDir
	ListDirs *ListDirs
}

type ResponseHeader struct {
	Error   *string `json:",omitempty"`
	Sending *int    `json:",omitempty"`
}

func (h *ResponseHeader) OK() bool {
	return h.Error == nil
}

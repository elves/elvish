package api

type Ping struct{}

type ListDirs struct {
	Blacklist map[string]struct{}
}

type Request struct {
	Ping     *Ping
	ListDirs *ListDirs
}

type ResponseHeader struct {
	Error   *string `json:",omitempty"`
	Sending *int    `json:",omitempty"`
}

func (h *ResponseHeader) OK() bool {
	return h.Error == nil
}

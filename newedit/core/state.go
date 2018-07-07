package core

type State struct {
	Mode    Mode
	Code    string
	Dot     int
	Pending *PendingCode
	Notes   []string
}

type PendingCode struct {
	Begin int
	End   int
	Text  string
}

func getMode(m Mode) Mode {
	if m == nil {
		return basicMode{}
	}
	return m
}

func (st *State) final() *State {
	return &State{Mode: basicMode{}, Code: st.Code, Dot: len(st.Code)}
}

func (st *State) CodeBeforeDot() string {
	return st.Code[:st.Dot]
}

func (st *State) CodeAfterDot() string {
	return st.Code[st.Dot:]
}

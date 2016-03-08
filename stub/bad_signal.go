package stub

type BadSignal struct {
	Error error
}

func (bs BadSignal) String() string {
	return "bad signal: " + bs.Error.Error()
}

func (BadSignal) Signal() {}

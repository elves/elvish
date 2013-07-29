package main

type Res struct {
	Cmd *ResCmd
	ProcState *ResProcState
}

type ResCmd struct {
	Pid int
}

type ResProcState struct {
	Pid int
	Exited bool
	ExitStatus int
	Signaled bool
	TermSig int
	CoreDump bool
	Stopped bool
	StopSig int
	Continued bool
}

func RecvRes() (r Res, err error) {
	err = ResDecoder.Decode(&r)
	return
}

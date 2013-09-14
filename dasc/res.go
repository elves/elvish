package main

type Res struct {
	Cmd *ResCmd
	ProcState *ResProcState
	BadRequest *ResBadRequest
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

type ResBadRequest struct {
	Err string
}

func RecvRes() (r Res, err error) {
	err = ResDecoder.Decode(&r)
	return
}

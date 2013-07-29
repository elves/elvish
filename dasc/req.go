package main

import (
	"fmt"
	"syscall"
	"encoding/json"
)

type ReqCmd struct {
	Path string
	Args []string
	Env map[string]string
	RedirOutput bool
	Output int `json:"-"`
}

type ReqExit struct {
}

type Req struct {
	Cmd *ReqCmd `json:",omitempty"`
	Exit *ReqExit `json:",omitempty"`
}

func sendFd(fd int) {
	cmsg := syscall.UnixRights(fd)
	err := syscall.Sendmsg(FdTube, nil, cmsg, nil, 0)
	if err != nil {
		fmt.Printf("Failed to sendmsg: %v\n", err)
	}
}

func SendReq(req Req) {
	json, err := json.Marshal(req)
	if err != nil {
		panic("failed to marshal request")
	}
	TextTube.Write(json)
	TextTube.WriteString("\n")

	cmd := req.Cmd
	if cmd != nil {
		if cmd.RedirOutput {
			sendFd(cmd.Output)
		}
	}
}

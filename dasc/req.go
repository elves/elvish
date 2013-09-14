package main

import (
	"fmt"
	"syscall"
	"encoding/json"
)

// Special fds used in Redirs[i][1]. These are all negative.
const (
	FdClose int = -iota-1
	FdSend
)

type ReqCmd struct {
	Path string
	Args []string
	Env map[string]string
	// A list of [oldfd, newfd] tuples. newfd may take negative special
	// values.
	Redirs [][2]int
	FdsToSend []int `json:"-"`
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
		for i, r := range cmd.Redirs {
			if r[1] == FdSend {
				sendFd(cmd.FdsToSend[i])
			}
		}
	}
}

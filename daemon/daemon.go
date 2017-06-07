// Package daemon implements a daemon for mediating access to the storage
// backend of elvish.
package daemon

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[daemon] ")

// Daemon is a daemon.
type Daemon struct {
	sockpath string
	dbpath   string
}

// New creates a new daemon.
func New(sockpath, dbpath string) *Daemon {
	return &Daemon{sockpath, dbpath}
}

// Main runs the daemon. It does not take care of forking and stuff; it assumes
// that it is already running in the correct process.
func (d *Daemon) Main() int {
	logger.Println("pid is", syscall.Getpid())

	st, err := store.NewStore(d.dbpath)
	if err != nil {
		logger.Print(err)
		return 2
	}

	listener, err := net.Listen("unix", d.sockpath)
	if err != nil {
		logger.Println("listen:", err)
		return 2
	}
	defer os.Remove(d.sockpath)

	cancel := make(chan struct{})
	sigterm := make(chan os.Signal)
	signal.Notify(sigterm, syscall.SIGTERM)
	go func() {
		<-sigterm
		close(cancel)
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-cancel:
				return 0
			default:
				logger.Println("accept:", err)
				return 2
			}
		}
		go handle(conn, st, cancel)
	}
}

func handle(c net.Conn, st *store.Store, cancel <-chan struct{}) {
	defer c.Close()
	go func() {
		<-cancel
		c.Close()
	}()

	decoder := json.NewDecoder(c)
	encoder := json.NewEncoder(c)

	send := func(v interface{}) {
		err := encoder.Encode(v)
		if err != nil {
			logger.Println("send:", err)
		}
	}
	sendOKHeader := func(n int) {
		send(&api.ResponseHeader{Sending: &n})
	}
	sendErrorHeader := func(e string) {
		send(&api.ResponseHeader{Error: &e})
	}

	for {
		var req api.Request
		err := decoder.Decode(&req)
		if err == io.EOF {
			return
		}
		if err != nil {
			sendErrorHeader("decode: " + err.Error())
			return
		}
		switch {
		case req.GetPid != nil:
			sendOKHeader(1)
			send(syscall.Getpid())
		case req.NextCmdSeq != nil:
			seq, err := st.NextCmdSeq()
			if err != nil {
				sendErrorHeader("NextCmdSeq: " + err.Error())
			} else {
				sendOKHeader(1)
				send(seq)
			}
		case req.AddCmd != nil:
			_, err := st.AddCmd(req.AddCmd.Text)
			if err != nil {
				sendErrorHeader("AddCmd: " + err.Error())
			} else {
				sendOKHeader(0)
			}
		case req.GetCmds != nil:
			// TODO: stream from store
			cmds, err := st.Cmds(req.GetCmds.From, req.GetCmds.Upto)
			if err != nil {
				sendErrorHeader("GetCmds: " + err.Error())
			} else {
				sendOKHeader(len(cmds))
				for _, cmd := range cmds {
					send(cmd)
				}
			}
		case req.AddDir != nil:
			err := st.AddDir(req.AddDir.Dir, req.AddDir.IncFactor)
			if err != nil {
				sendErrorHeader("AddDir: " + err.Error())
			} else {
				sendOKHeader(0)
			}
		case req.GetDirs != nil:
			dirs, err := st.GetDirs(req.GetDirs.Blacklist)
			if err != nil {
				sendErrorHeader("ListDirs: " + err.Error())
			} else {
				sendOKHeader(len(dirs))
				for _, dir := range dirs {
					send(dir)
				}
			}
		case req.GetSharedVar != nil:
			value, err := st.GetSharedVar(req.GetSharedVar.Name)
			if err != nil {
				sendErrorHeader("GetSharedVar: " + err.Error())
			} else {
				sendOKHeader(1)
				send(value)
			}
		case req.SetSharedVar != nil:
			r := req.SetSharedVar
			err := st.SetSharedVar(r.Name, r.Value)
			if err != nil {
				sendErrorHeader("SetSharedVar: " + err.Error())
			} else {
				sendOKHeader(0)
			}
		case req.DelSharedVar != nil:
			err := st.DelSharedVar(req.DelSharedVar.Name)
			if err != nil {
				sendErrorHeader("DelSharedVar: " + err.Error())
			} else {
				sendOKHeader(0)
			}
		default:
			sendErrorHeader("bad request")
		}
	}
}

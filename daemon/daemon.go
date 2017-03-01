// Package elvishd implements a daemon for mediating access to the storage backend of elvish.
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

var Logger = util.GetLogger("[daemon] ")

type Daemon struct {
	sockpath string
	dbpath   string
}

func New(sockpath, dbpath string) *Daemon {
	return &Daemon{sockpath, dbpath}
}

func (d *Daemon) Main() int {
	Logger.Println("pid is", syscall.Getpid())

	st, err := store.NewStore(d.dbpath)
	if err != nil {
		Logger.Print(err)
		return 2
	}

	listener, err := net.Listen("unix", d.sockpath)
	if err != nil {
		Logger.Println("listen:", err)
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
				Logger.Println("accept:", err)
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
			Logger.Println("send:", err)
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
		case req.AddDir != nil:
			err := st.AddDir(req.AddDir.Dir, req.AddDir.IncFactor)
			if err != nil {
				sendErrorHeader("AddDir: " + err.Error())
			} else {
				sendOKHeader(0)
			}
		case req.ListDirs != nil:
			dirs, err := st.GetDirs(req.ListDirs.Blacklist)
			if err != nil {
				sendErrorHeader("ListDirs: " + err.Error())
				continue
			}
			sendOKHeader(len(dirs))
			for _, dir := range dirs {
				send(dir)
			}
		// case req.Quit:
		default:
			sendErrorHeader("bad request")
		}
	}
}

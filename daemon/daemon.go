// Package elvishd implements a daemon for mediating access to the storage backend of elvish.
package daemon

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/store"
)

type Daemon struct {
	sockpath string
	dbpath   string
}

func New(sockpath, dbpath string) *Daemon {
	return &Daemon{sockpath, dbpath}
}

func (d *Daemon) Main() int {
	st, err := store.NewStore(d.dbpath)
	if err != nil {
		log.Print(err)
		return 2
	}

	listener, err := net.Listen("unix", d.sockpath)
	if err != nil {
		log.Println("listen:", err)
		return 2
	}
	defer os.Remove(d.sockpath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept:", err)
			return 2
		}
		go handle(conn, st)
	}
}

func handle(c net.Conn, st *store.Store) {
	defer c.Close()
	decoder := json.NewDecoder(c)
	encoder := json.NewEncoder(c)

	send := func(v interface{}) {
		err := encoder.Encode(v)
		if err != nil {
			log.Println("send:", err)
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
		case req.Ping != nil:
			sendOKHeader(0)
		case req.ListDirs != nil:
			dirs, err := st.ListDirs(req.ListDirs.Blacklist)
			if err != nil {
				sendErrorHeader("listdir: " + err.Error())
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

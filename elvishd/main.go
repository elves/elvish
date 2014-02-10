// elvishd is an agent for sharing variables and command-line history among
// multiple elvish processes.
package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path"
	"syscall"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xiaq/elvish/service"
	"github.com/xiaq/elvish/util"
)

const (
	SignalBufferSize = 32
)

func main() {
	laddr, err := util.SocketName()
	if err != nil {
		log.Fatalln("get socket name:", err)
	}

	// Listen to socket
	listener, err := net.Listen("unix", laddr)
	if err != nil {
		log.Fatalln("listen to socket:", err)
	}

	// Construct database filename
	u, err := user.Current()
	if err != nil {
		log.Fatalln("get current user:", err)
	}
	home := u.HomeDir
	if home == "" {
		log.Fatalln("current user does not have a home directory")
	}
	dbname := path.Join(home, ".elvishd.db")

	// Open database and construct dbmap
	db, err := sql.Open("sqlite3", dbname)
	if err != nil {
		log.Fatalln("open database:", err)
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// Set up Unix signal handler
	sigch := make(chan os.Signal, SignalBufferSize)
	signal.Notify(sigch)
	go func() {
		for sig := range sigch {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				// TODO(xiaq): Notify current clients of termination
				os.Remove(laddr)
				db.Close() // Ignore possible errors
				os.Exit(0)
			default:
				// Ignore all other signals
			}
		}
	}()

	err = service.Serve(listener, dbmap)
	if err != nil {
		log.Fatalln("start service:", err)
	}
}

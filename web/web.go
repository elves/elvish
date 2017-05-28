// Package web is the entry point for the backend of the web interface of
// Elvish.
package web

//go:generate ./embed-html

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store"
)

type Web struct {
	ev   *eval.Evaler
	st   *store.Store
	port int
}

func NewWeb(ev *eval.Evaler, st *store.Store, port int) *Web {
	return &Web{ev, st, port}
}

func (web *Web) Run(args []string) int {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "arguments to -web are not supported yet")
		return 2
	}

	http.HandleFunc("/", web.handleMainPage)
	http.HandleFunc("/execute", web.handleExecute)
	addr := fmt.Sprintf("localhost:%d", web.port)
	log.Println("going to listen", addr)
	err := http.ListenAndServe(addr, nil)

	log.Println(err)
	return 0
}

func (web *Web) handleMainPage(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(mainPageHTML))
	if err != nil {
		log.Println("cannot write response:", err)
	}
}

func (web *Web) handleExecute(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("response!"))
	if err != nil {
		log.Println("cannot write response:", err)
	}
}

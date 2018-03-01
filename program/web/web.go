// Package web is the entry point for the backend of the web interface of
// Elvish.
package web

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/runtime"
)

type Web struct {
	BinPath  string
	SockPath string
	DbPath   string
	Port     int
}

type httpHandler struct {
	ev *eval.Evaler
}

type ExecuteResponse struct {
	OutBytes  string
	OutValues []interface{}
	ErrBytes  string
	Err       string
}

func New(binpath, sockpath, dbpath string, port int) *Web {
	return &Web{binpath, sockpath, dbpath, port}
}

func (web *Web) Main([]string) int {
	ev, _ := runtime.InitRuntime(web.BinPath, web.SockPath, web.DbPath)
	defer runtime.CleanupRuntime(ev)

	h := httpHandler{ev}

	http.HandleFunc("/", h.handleMainPage)
	http.HandleFunc("/execute", h.handleExecute)
	addr := fmt.Sprintf("localhost:%d", web.Port)
	log.Println("going to listen", addr)
	err := http.ListenAndServe(addr, nil)

	log.Println(err)
	return 0
}

func (h httpHandler) handleMainPage(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(mainPageHTML))
	if err != nil {
		log.Println("cannot write response:", err)
	}
}

func (h httpHandler) handleExecute(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("cannot read request body:", err)
		return
	}
	code := string(bytes)

	outBytes, outValues, errBytes, err := evalAndCollect(h.ev, code)
	errText := ""
	if err != nil {
		errText = err.Error()
	}
	responseBody, err := json.Marshal(
		&ExecuteResponse{string(outBytes), outValues, string(errBytes), errText})
	if err != nil {
		log.Println("cannot marshal response body:", err)
	}

	_, err = w.Write(responseBody)
	if err != nil {
		log.Println("cannot write response:", err)
	}
}

const (
	outFileBufferSize = 1024
	outChanBufferSize = 32
)

// evalAndCollect evaluates a piece of code with null stdin, and stdout and
// stderr connected to pipes (value part of stderr being a blackhole), and
// return the results collected on stdout and stderr, and the possible error
// that occurred.
func evalAndCollect(ev *eval.Evaler, code string) (
	outBytes []byte, outValues []interface{}, errBytes []byte, err error) {

	node, err := parse.Parse("[web]", code)
	if err != nil {
		return
	}
	src := eval.NewInteractiveSource(code)
	op, err := ev.Compile(node, src)
	if err != nil {
		return
	}

	outFile, chanOutBytes := makeBytesWriterAndCollect()
	outChan, chanOutValues := makeValuesWriterAndCollect()
	errFile, chanErrBytes := makeBytesWriterAndCollect()

	ports := []*eval.Port{
		eval.DevNullClosedChan,
		{File: outFile, Chan: outChan},
		{File: errFile, Chan: eval.BlackholeChan},
	}
	err = ev.Eval(op, ports, src)

	outFile.Close()
	close(outChan)
	errFile.Close()
	return <-chanOutBytes, <-chanOutValues, <-chanErrBytes, err
}

// makeBytesWriterAndCollect makes an in-memory file that can be written to, and
// the written bytes will be collected in a byte slice that will be put on a
// channel as soon as the writer is closed.
func makeBytesWriterAndCollect() (*os.File, <-chan []byte) {
	r, w, err := os.Pipe()
	// os.Pipe returns error only on resource exhaustion.
	if err != nil {
		panic(err)
	}
	chanCollected := make(chan []byte)

	go func() {
		var (
			collected []byte
			buf       [outFileBufferSize]byte
		)
		for {
			n, err := r.Read(buf[:])
			collected = append(collected, buf[:n]...)
			if err != nil {
				if err != io.EOF {
					log.Println("error when reading output pipe:", err)
				}
				break
			}
		}
		r.Close()
		chanCollected <- collected
	}()

	return w, chanCollected
}

// makeValuesWriterAndCollect makes a Value channel for writing, and the written
// values will be collected in a Value slice that will be put on a channel as
// soon as the writer is closed.
func makeValuesWriterAndCollect() (chan interface{}, <-chan []interface{}) {
	chanValues := make(chan interface{}, outChanBufferSize)
	chanCollected := make(chan []interface{})

	go func() {
		var collected []interface{}
		for {
			for v := range chanValues {
				collected = append(collected, v)
			}
			chanCollected <- collected
		}
	}()

	return chanValues, chanCollected
}

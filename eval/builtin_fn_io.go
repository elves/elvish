package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/elves/elvish/eval/vals"
)

// Input and output.

func init() {
	addBuiltinFns(map[string]interface{}{
		// Value output
		"put": put,

		// Bytes output
		"print":  print,
		"echo":   echo,
		"pprint": pprint,
		"repr":   repr,

		// Only bytes or values
		//
		// These are now implemented as commands forwarding one part of input to
		// output and discarding the other. A future optimization the evaler can
		// do is to connect the relevant parts directly together without any
		// kind of forwarding.
		"only-bytes":  onlyBytes,
		"only-values": onlyValues,

		// Bytes to value
		"slurp":      slurp,
		"from-lines": fromLines,
		"from-json":  fromJSON,

		// Value to bytes
		"to-lines": toLines,
		"to-json":  toJSON,

		// File and pipe
		"fopen":   fopen,
		"fclose":  fclose,
		"pipe":    pipe,
		"prclose": prclose,
		"pwclose": pwclose,
	})
}

func put(fm *Frame, args ...interface{}) {
	out := fm.ports[1].Chan
	for _, a := range args {
		out <- a
	}
}

type printOpts struct{ Sep string }

func (o *printOpts) SetDefaultOptions() { o.Sep = " " }

func print(fm *Frame, opts printOpts, args ...interface{}) {
	out := fm.ports[1].File
	for i, arg := range args {
		if i > 0 {
			out.WriteString(opts.Sep)
		}
		out.WriteString(vals.ToString(arg))
	}
}

func echo(fm *Frame, opts printOpts, args ...interface{}) {
	print(fm, opts, args...)
	fm.ports[1].File.WriteString("\n")
}

func pprint(fm *Frame, args ...interface{}) {
	out := fm.ports[1].File
	for _, arg := range args {
		out.WriteString(vals.Repr(arg, 0))
		out.WriteString("\n")
	}
}

func repr(fm *Frame, args ...interface{}) {
	out := fm.ports[1].File
	for i, arg := range args {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(vals.Repr(arg, vals.NoPretty))
	}
	out.WriteString("\n")
}

const bytesReadBufferSize = 512

func onlyBytes(fm *Frame) error {
	// Discard values in a goroutine.
	valuesDone := make(chan struct{})
	go func() {
		for range fm.InputChan() {
		}
		close(valuesDone)
	}()
	// Make sure the goroutine has finished before returning.
	defer func() { <-valuesDone }()

	// Forward bytes.
	buf := make([]byte, bytesReadBufferSize)
	for {
		nr, errRead := fm.InputFile().Read(buf[:])
		if nr > 0 {
			// Even when there are write errors, we will continue reading. So we
			// ignore the error.
			fm.OutputFile().Write(buf[:nr])
		}
		if errRead != nil {
			if errRead == io.EOF {
				return nil
			}
			return errRead
		}
	}
}

func onlyValues(fm *Frame) error {
	// Forward values in a goroutine.
	valuesDone := make(chan struct{})
	go func() {
		for v := range fm.InputChan() {
			fm.OutputChan() <- v
		}
		close(valuesDone)
	}()
	// Make sure the goroutine has finished before returning.
	defer func() { <-valuesDone }()

	// Discard bytes.
	buf := make([]byte, bytesReadBufferSize)
	for {
		_, errRead := fm.InputFile().Read(buf[:])
		if errRead != nil {
			if errRead == io.EOF {
				return nil
			}
			return errRead
		}
	}
}

func slurp(fm *Frame) (string, error) {
	b, err := ioutil.ReadAll(fm.ports[0].File)
	return string(b), err
}

func fromLines(fm *Frame) {
	linesToChan(fm.ports[0].File, fm.ports[1].Chan)
}

// fromJSON parses a stream of JSON data into Value's.
func fromJSON(fm *Frame) error {
	in := fm.ports[0].File
	out := fm.ports[1].Chan

	dec := json.NewDecoder(in)
	for {
		var v interface{}
		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		converted, err := fromJSONInterface(v)
		if err != nil {
			return err
		}
		out <- converted
	}
}

func toLines(fm *Frame, inputs Inputs) {
	out := fm.ports[1].File

	inputs(func(v interface{}) {
		fmt.Fprintln(out, vals.ToString(v))
	})
}

// toJSON converts a stream of Value's to JSON data.
func toJSON(fm *Frame, inputs Inputs) error {
	encoder := json.NewEncoder(fm.OutputFile())

	var errEncode error
	inputs(func(v interface{}) {
		if errEncode != nil {
			return
		}
		errEncode = encoder.Encode(v)
	})
	return errEncode
}

func fopen(fm *Frame, name string) (vals.File, error) {
	// TODO support opening files for writing etc as well.
	f, err := os.Open(name)
	return f, err
}

func fclose(f vals.File) error {
	return f.Close()
}

func pipe() (vals.Pipe, error) {
	r, w, err := os.Pipe()
	return vals.NewPipe(r, w), err
}

func prclose(p vals.Pipe) error {
	return p.ReadEnd.Close()
}

func pwclose(p vals.Pipe) error {
	return p.WriteEnd.Close()
}

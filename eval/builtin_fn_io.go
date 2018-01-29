package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/elves/elvish/eval/types"
)

// Input and output.

func init() {
	addToBuiltinFns([]*BuiltinFn{
		// Value output
		{"put", put},

		// Bytes output
		{"print", print},
		{"echo", echo},
		{"pprint", pprint},
		{"repr", repr},

		// Bytes to value
		{"slurp", slurp},
		{"from-lines", fromLines},
		{"from-json", fromJSON},

		// Value to bytes
		{"to-lines", toLines},
		{"to-json", toJSON},

		// File and pipe
		{"fopen", fopen},
		{"fclose", fclose},
		{"pipe", pipe},
		{"prclose", prclose},
		{"pwclose", pwclose},
	})
}

func put(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- a
	}
}

func print(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var sepv string
	ScanOpts(opts, OptToScan{"sep", &sepv, " "})

	out := ec.ports[1].File
	sep := sepv
	for i, arg := range args {
		if i > 0 {
			out.WriteString(sep)
		}
		out.WriteString(types.ToString(arg))
	}
}

func echo(ec *Frame, args []interface{}, opts map[string]interface{}) {
	print(ec, args, opts)
	ec.ports[1].File.WriteString("\n")
}

func pprint(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	out := ec.ports[1].File
	for _, arg := range args {
		out.WriteString(types.Repr(arg, 0))
		out.WriteString("\n")
	}
}

func repr(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	out := ec.ports[1].File
	for i, arg := range args {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(types.Repr(arg, types.NoPretty))
	}
	out.WriteString("\n")
}

func slurp(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	in := ec.ports[0].File
	out := ec.ports[1].Chan

	all, err := ioutil.ReadAll(in)
	maybeThrow(err)
	out <- string(all)
}

func fromLines(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	in := ec.ports[0].File
	out := ec.ports[1].Chan

	linesToChan(in, out)
}

// fromJSON parses a stream of JSON data into Value's.
func fromJSON(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	in := ec.ports[0].File
	out := ec.ports[1].Chan

	dec := json.NewDecoder(in)
	var v interface{}
	for {
		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				return
			}
			throw(err)
		}
		out <- FromJSONInterface(v)
	}
}

func toLines(ec *Frame, args []interface{}, opts map[string]interface{}) {
	iterate := ScanArgsOptionalInput(ec, args)
	TakeNoOpt(opts)

	out := ec.ports[1].File

	iterate(func(v interface{}) {
		fmt.Fprintln(out, types.ToString(v))
	})
}

// toJSON converts a stream of Value's to JSON data.
func toJSON(ec *Frame, args []interface{}, opts map[string]interface{}) {
	iterate := ScanArgsOptionalInput(ec, args)
	TakeNoOpt(opts)

	out := ec.ports[1].File

	enc := json.NewEncoder(out)
	iterate(func(v interface{}) {
		err := enc.Encode(v)
		maybeThrow(err)
	})
}

func fopen(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var namev string
	ScanArgs(args, &namev)
	name := namev
	TakeNoOpt(opts)

	// TODO support opening files for writing etc as well.
	out := ec.ports[1].Chan
	f, err := os.Open(name)
	maybeThrow(err)
	out <- types.File{f}
}

func fclose(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var f types.File
	ScanArgs(args, &f)
	TakeNoOpt(opts)

	maybeThrow(f.Inner.Close())
}

func pipe(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	r, w, err := os.Pipe()
	out := ec.ports[1].Chan
	maybeThrow(err)
	out <- types.Pipe{r, w}
}

func prclose(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var p types.Pipe
	ScanArgs(args, &p)
	TakeNoOpt(opts)

	maybeThrow(p.ReadEnd.Close())
}

func pwclose(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var p types.Pipe
	ScanArgs(args, &p)
	TakeNoOpt(opts)

	maybeThrow(p.WriteEnd.Close())
}

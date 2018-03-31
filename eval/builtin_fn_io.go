package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/elves/elvish/eval/vals"
	"github.com/xiaq/persistent/hashmap"
	"github.com/xiaq/persistent/vector"
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

func print(fm *Frame, rawOpts RawOptions, args ...interface{}) {
	opts := struct{ Sep string }{" "}
	rawOpts.Scan(&opts)

	out := fm.ports[1].File
	for i, arg := range args {
		if i > 0 {
			out.WriteString(opts.Sep)
		}
		out.WriteString(vals.ToString(arg))
	}
}

func echo(fm *Frame, opts RawOptions, args ...interface{}) {
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
	var v interface{}
	for {
		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		out <- FromJSONInterface(v)
	}
	return nil
}

func toLines(fm *Frame, inputs Inputs) {
	out := fm.ports[1].File

	inputs(func(v interface{}) {
		fmt.Fprintln(out, vals.ToString(v))
	})
}

// Exported so that other packages (namely, ones that implement an
// elvish namespace that provide Go functions that deal with var ->
// JSON transformation) can refer to it for numberification.
func FixVarForJSON(src interface{}, numberify bool) interface{} {
	if !numberify {
		return src
	}
	switch v := src.(type) {
	case string:
		if numberify {
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			} else if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
		return src
	case vector.Vector:
		res := vals.EmptyList
		for it := v.Iterator(); it.HasElem(); it.Next() {
			res = res.Cons(FixVarForJSON(it.Elem(), numberify))
		}
		return res
	case hashmap.Map:
		res := vals.EmptyMap
		for it := v.Iterator(); it.HasElem(); it.Next() {
			key, elem := it.Elem()
			res = res.Assoc(key, FixVarForJSON(elem, numberify))
		}
		return res
	default:
		return src
	}
}

// toJSON converts a stream of Value's to JSON data.
func toJSON(fm *Frame, rawOpts RawOptions, inputs Inputs) {
	opts := struct {
		Numberify bool `name:"-numberify"`
	}{false}
	rawOpts.Scan(&opts)
	out := fm.ports[1].File

	enc := json.NewEncoder(out)
	inputs(func(v interface{}) {
		err := enc.Encode(FixVarForJSON(v, opts.Numberify))
		maybeThrow(err)
	})
}

func fopen(fm *Frame, name string) (vals.File, error) {
	// TODO support opening files for writing etc as well.
	f, err := os.Open(name)
	return vals.File{f}, err
}

func fclose(f vals.File) error {
	return f.Inner.Close()
}

func pipe() (vals.Pipe, error) {
	r, w, err := os.Pipe()
	return vals.Pipe{r, w}, err
}

func prclose(p vals.Pipe) error {
	return p.ReadEnd.Close()
}

func pwclose(p vals.Pipe) error {
	return p.WriteEnd.Close()
}

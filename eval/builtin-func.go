package eval

// Builtin functions.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
)

type builtinFuncImpl func(*Evaluator, []Value) Exitus

type builtinFunc struct {
	fn          builtinFuncImpl
	streamTypes [2]StreamType
}

var builtinFuncs map[string]builtinFunc

func init() {
	// Needed to work around init loop.
	builtinFuncs = map[string]builtinFunc{
		"print":   builtinFunc{print, [2]StreamType{0, fdStream}},
		"println": builtinFunc{println, [2]StreamType{0, fdStream}},

		"printchan": builtinFunc{printchan, [2]StreamType{chanStream, fdStream}},
		"feedchan":  builtinFunc{feedchan, [2]StreamType{fdStream, chanStream}},

		"put":    builtinFunc{put, [2]StreamType{0, chanStream}},
		"unpack": builtinFunc{unpack, [2]StreamType{chanStream, chanStream}},

		"parse-json": builtinFunc{parseJSON, [2]StreamType{fdStream, chanStream}},

		"typeof": builtinFunc{typeof, [2]StreamType{0, chanStream}},

		"failure": builtinFunc{failure, [2]StreamType{0, chanStream}},

		"each": builtinFunc{each, [2]StreamType{chanStream, hybridStream}},

		"if": builtinFunc{ifFn, [2]StreamType{hybridStream, hybridStream}},

		"cd": builtinFunc{cd, [2]StreamType{}},

		"source": builtinFunc{source, [2]StreamType{hybridStream, hybridStream}},

		"+": builtinFunc{plus, [2]StreamType{0, chanStream}},
		"-": builtinFunc{minus, [2]StreamType{0, chanStream}},
		"*": builtinFunc{times, [2]StreamType{0, chanStream}},
		"/": builtinFunc{divide, [2]StreamType{0, chanStream}},

		"=": builtinFunc{eq, [2]StreamType{0, chanStream}},
	}
}

var (
	argsError  = newFailure("args error")
	inputError = newFailure("input error")
)

func put(ev *Evaluator, args []Value) Exitus {
	out := ev.ports[1].ch
	for _, a := range args {
		out <- a
	}
	return success
}

func typeof(ev *Evaluator, args []Value) Exitus {
	out := ev.ports[1].ch
	for _, a := range args {
		out <- NewString(a.Type().String())
	}
	return success
}

func failure(ev *Evaluator, args []Value) Exitus {
	if len(args) != 1 {
		return argsError
	}
	out := ev.ports[1].ch
	out <- newFailure(args[0].String())
	return success
}

func print(ev *Evaluator, args []Value) Exitus {
	out := ev.ports[1].f
	for _, a := range args {
		fmt.Fprint(out, a.String())
	}
	return success
}

func println(ev *Evaluator, args []Value) Exitus {
	args = append(args, NewString("\n"))
	return print(ev, args)
}

func printchan(ev *Evaluator, args []Value) Exitus {
	if len(args) > 0 {
		return argsError
	}
	in := ev.ports[0].ch
	out := ev.ports[1].f

	for s := range in {
		fmt.Fprintln(out, s.String())
	}
	return success
}

func feedchan(ev *Evaluator, args []Value) Exitus {
	if len(args) > 0 {
		return argsError
	}
	in := ev.ports[0].f
	out := ev.ports[1].ch

	fmt.Println("WARNING: Only string input is supported at the moment.")

	bufferedIn := bufio.NewReader(in)
	// i := 0
	for {
		// fmt.Printf("[%v] ", i)
		line, err := bufferedIn.ReadString('\n')
		if err == io.EOF {
			return success
		} else if err != nil {
			return newFailure(err.Error())
		}
		out <- NewString(line[:len(line)-1])
		// i++
	}
}

// unpack takes any number of tables and output their list elements.
func unpack(ev *Evaluator, args []Value) Exitus {
	if len(args) != 0 {
		return argsError
	}
	in := ev.ports[0].ch
	out := ev.ports[1].ch

	for v := range in {
		if t, ok := v.(*Table); !ok {
			return inputError
		} else {
			for _, e := range t.List {
				out <- e
			}
		}
	}

	return success
}

// parseJSON parses a stream of JSON data into Value's.
func parseJSON(ev *Evaluator, args []Value) Exitus {
	if len(args) > 0 {
		return argsError
	}
	in := ev.ports[0].f
	out := ev.ports[1].ch

	dec := json.NewDecoder(in)
	var v interface{}
	for {
		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				return success
			} else {
				return newFailure(err.Error())
			}
		}
		out <- fromJSONInterface(v)
	}
}

// each takes a single closure and applies it to all input values.
func each(ev *Evaluator, args []Value) Exitus {
	if len(args) != 1 {
		return argsError
	}
	if f, ok := args[0].(*Closure); !ok {
		return argsError
	} else {
		in := ev.ports[0].ch
		for v := range in {
			newEv := ev.copy("closure of each")
			su := newEv.execClosure(f, []Value{v})
			for _ = range su {
			}
		}
	}
	return success
}

// if takes a sequence of values and a trailing nullary closure. If all of the
// values are true, the closure is executed.
func ifFn(ev *Evaluator, args []Value) Exitus {
	if len(args) == 0 {
		return argsError
	}
	if f, ok := args[len(args)-1].(*Closure); !ok {
		return argsError
	} else if len(f.ArgNames) > 0 {
		return argsError
	} else {
		for _, a := range args[:len(args)-1] {
			if !a.Bool() {
				return success
			}
		}
		newEv := ev.copy("closure of if")
		su := newEv.execClosure(f, []Value{})
		for _ = range su {
		}
		return success
	}
}

func cd(ev *Evaluator, args []Value) Exitus {
	var dir string
	if len(args) == 0 {
		user, err := user.Current()
		if err == nil {
			dir = user.HomeDir
		}
	} else if len(args) == 1 {
		dir = args[0].String()
	} else {
		return argsError
	}
	err := os.Chdir(dir)
	if err != nil {
		return newFailure(err.Error())
	}
	return success
}

func source(ev *Evaluator, args []Value) Exitus {
	if len(args) != 1 {
		return argsError
	}
	if fname, ok := args[0].(String); !ok {
		return argsError
	} else {
		ev.Source(string(fname))
	}
	return success
}

func toFloats(args []Value) (nums []float64, err error) {
	for _, a := range args {
		a, ok := a.(String)
		if !ok {
			return nil, fmt.Errorf("must be string")
		}
		f, err := strconv.ParseFloat(string(a), 64)
		if err != nil {
			return nil, err
		}
		nums = append(nums, f)
	}
	return
}

func plus(ev *Evaluator, args []Value) Exitus {
	out := ev.ports[1].ch
	nums, err := toFloats(args)
	if err != nil {
		return newFailure(err.Error())
	}
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	out <- NewString(fmt.Sprintf("%g", sum))
	return success
}

func minus(ev *Evaluator, args []Value) Exitus {
	out := ev.ports[1].ch
	if len(args) == 0 {
		return argsError
	}
	nums, err := toFloats(args)
	if err != nil {
		return newFailure(err.Error())
	}
	sum := nums[0]
	for _, f := range nums[1:] {
		sum -= f
	}
	out <- NewString(fmt.Sprintf("%g", sum))
	return success
}

func times(ev *Evaluator, args []Value) Exitus {
	out := ev.ports[1].ch
	nums, err := toFloats(args)
	if err != nil {
		return newFailure(err.Error())
	}
	prod := 1.0
	for _, f := range nums {
		prod *= f
	}
	out <- NewString(fmt.Sprintf("%g", prod))
	return success
}

func divide(ev *Evaluator, args []Value) Exitus {
	out := ev.ports[1].ch
	if len(args) == 0 {
		return argsError
	}
	nums, err := toFloats(args)
	if err != nil {
		return newFailure(err.Error())
	}
	prod := nums[0]
	for _, f := range nums[1:] {
		prod /= f
	}
	out <- NewString(fmt.Sprintf("%g", prod))
	return success
}

func eq(ev *Evaluator, args []Value) Exitus {
	out := ev.ports[1].ch
	if len(args) == 0 {
		return argsError
	}
	for i := 0; i+1 < len(args); i++ {
		if !valueEq(args[i], args[i+1]) {
			out <- Bool(false)
			return success
		}
	}
	out <- Bool(true)
	return success
}

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

var builtinFns []*builtinFn

func init() {
	// Needed to work around init loop.
	builtinFns = []*builtinFn{
		&builtinFn{"print", print},
		&builtinFn{"println", println},

		&builtinFn{"printchan", printchan},
		&builtinFn{"feedchan", feedchan},

		&builtinFn{"rat", ratFn},

		&builtinFn{"put", put},
		&builtinFn{"unpack", unpack},

		&builtinFn{"parse-json", parseJSON},

		&builtinFn{"typeof", typeof},

		&builtinFn{"failure", failure},

		&builtinFn{"each", each},

		&builtinFn{"cd", cd},
		&builtinFn{"visited-dirs", visistedDirs},
		&builtinFn{"jump-dir", jumpDir},

		&builtinFn{"source", source},

		&builtinFn{"+", plus},
		&builtinFn{"-", minus},
		&builtinFn{"*", times},
		&builtinFn{"/", divide},

		&builtinFn{"=", eq},
	}
}

var (
	argsError  = newFailure("args error")
	inputError = newFailure("input error")
)

func put(ev *Evaluator, args []Value) exitus {
	out := ev.ports[1].ch
	for _, a := range args {
		out <- a
	}
	return success
}

func typeof(ev *Evaluator, args []Value) exitus {
	out := ev.ports[1].ch
	for _, a := range args {
		out <- str(a.Type().String())
	}
	return success
}

func failure(ev *Evaluator, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	out := ev.ports[1].ch
	out <- newFailure(toString(args[0]))
	return success
}

func print(ev *Evaluator, args []Value) exitus {
	out := ev.ports[1].f
	for _, a := range args {
		fmt.Fprint(out, toString(a))
	}
	return success
}

func println(ev *Evaluator, args []Value) exitus {
	args = append(args, str("\n"))
	return print(ev, args)
}

func printchan(ev *Evaluator, args []Value) exitus {
	if len(args) > 0 {
		return argsError
	}
	in := ev.ports[0].ch
	out := ev.ports[1].f

	for v := range in {
		fmt.Fprintln(out, toString(v))
	}
	return success
}

func feedchan(ev *Evaluator, args []Value) exitus {
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
		out <- str(line[:len(line)-1])
		// i++
	}
}

func ratFn(ev *Evaluator, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	out := ev.ports[1].ch
	r, err := toRat(args[0])
	if err != nil {
		return newFailure(err.Error())
	}
	out <- r
	return success
}

// unpack takes any number of tables and output their list elements.
func unpack(ev *Evaluator, args []Value) exitus {
	if len(args) != 0 {
		return argsError
	}
	in := ev.ports[0].ch
	out := ev.ports[1].ch

	for v := range in {
		if t, ok := v.(*table); !ok {
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
func parseJSON(ev *Evaluator, args []Value) exitus {
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
			}
			return newFailure(err.Error())
		}
		out <- fromJSONInterface(v)
	}
}

// each takes a single closure and applies it to all input values.
func each(ev *Evaluator, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	if f, ok := args[0].(*closure); !ok {
		return argsError
	} else {
		in := ev.ports[0].ch
		for v := range in {
			su := f.Exec(ev.copy("closure of each"), []Value{v})
			for _ = range su {
			}
		}
	}
	return success
}

func cd(ev *Evaluator, args []Value) exitus {
	var dir string
	if len(args) == 0 {
		user, err := user.Current()
		if err == nil {
			dir = user.HomeDir
		}
	} else if len(args) == 1 {
		dir = toString(args[0])
	} else {
		return argsError
	}
	err := os.Chdir(dir)
	if err != nil {
		return newFailure(err.Error())
	}
	if ev.store != nil {
		pwd, err := os.Getwd()
		// XXX(xiaq): ignores error
		if err == nil {
			ev.store.AddVisitedDir(pwd)
		}
	}
	return success
}

var storeNotConnected = newFailure("store not connected")

func visistedDirs(ev *Evaluator, args []Value) exitus {
	if ev.store == nil {
		return storeNotConnected
	}
	dirs, err := ev.store.ListVisitedDirs()
	if err != nil {
		return newFailure("store error: " + err.Error())
	}
	out := ev.ports[1].ch
	for _, dir := range dirs {
		table := newTable()
		table.Dict["path"] = str(dir.Path)
		table.Dict["score"] = str(fmt.Sprint(dir.Score))
		out <- table
	}
	return success
}

var noMatchingDir = newFailure("no matching directory")

func jumpDir(ev *Evaluator, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	if ev.store == nil {
		return storeNotConnected
	}
	dirs, err := ev.store.FindVisitedDirs(toString(args[0]))
	if err != nil {
		return newFailure("store error: " + err.Error())
	}
	if len(dirs) == 0 {
		return noMatchingDir
	}
	dir := dirs[0].Path
	err = os.Chdir(dir)
	// TODO(xiaq): Remove directories that no longer exist
	if err != nil {
		return newFailure(err.Error())
	}
	ev.store.AddVisitedDir(dir)
	return success
}

func source(ev *Evaluator, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	if fname, ok := args[0].(str); !ok {
		return argsError
	} else {
		ev.Source(string(fname))
	}
	return success
}

func toFloats(args []Value) (nums []float64, err error) {
	for _, a := range args {
		a, ok := a.(str)
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

func plus(ev *Evaluator, args []Value) exitus {
	out := ev.ports[1].ch
	nums, err := toFloats(args)
	if err != nil {
		return newFailure(err.Error())
	}
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	out <- str(fmt.Sprintf("%g", sum))
	return success
}

func minus(ev *Evaluator, args []Value) exitus {
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
	out <- str(fmt.Sprintf("%g", sum))
	return success
}

func times(ev *Evaluator, args []Value) exitus {
	out := ev.ports[1].ch
	nums, err := toFloats(args)
	if err != nil {
		return newFailure(err.Error())
	}
	prod := 1.0
	for _, f := range nums {
		prod *= f
	}
	out <- str(fmt.Sprintf("%g", prod))
	return success
}

func divide(ev *Evaluator, args []Value) exitus {
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
	out <- str(fmt.Sprintf("%g", prod))
	return success
}

func eq(ev *Evaluator, args []Value) exitus {
	out := ev.ports[1].ch
	if len(args) == 0 {
		return argsError
	}
	for i := 0; i+1 < len(args); i++ {
		if !valueEq(args[i], args[i+1]) {
			out <- boolean(false)
			return success
		}
	}
	out <- boolean(true)
	return success
}

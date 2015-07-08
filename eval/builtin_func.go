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
var BuiltinFnNames []string

func init() {
	// Needed to work around init loop.
	builtinFns = []*builtinFn{
		&builtinFn{":", nop},

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
		&builtinFn{"return", returnFn},
		&builtinFn{"break", breakFn},
		&builtinFn{"continue", continueFn},

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
	for _, b := range builtinFns {
		BuiltinFnNames = append(BuiltinFnNames, b.Name)
	}
}

var (
	argsError  = newFailure("args error")
	inputError = newFailure("input error")
)

func nop(ec *evalCtx, args []Value) exitus {
	return success
}

func put(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
	for _, a := range args {
		out <- a
	}
	return success
}

func typeof(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
	for _, a := range args {
		out <- str(a.Type().String())
	}
	return success
}

func failure(ec *evalCtx, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	out := ec.ports[1].ch
	out <- newFailure(toString(args[0]))
	return success
}

func returnFn(ec *evalCtx, args []Value) exitus {
	return newFlowExitus(Return)
}

func breakFn(ec *evalCtx, args []Value) exitus {
	return newFlowExitus(Break)
}

func continueFn(ec *evalCtx, args []Value) exitus {
	return newFlowExitus(Continue)
}

func print(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].f
	for _, a := range args {
		fmt.Fprint(out, toString(a))
	}
	return success
}

func println(ec *evalCtx, args []Value) exitus {
	args = append(args, str("\n"))
	return print(ec, args)
}

func printchan(ec *evalCtx, args []Value) exitus {
	if len(args) > 0 {
		return argsError
	}
	in := ec.ports[0].ch
	out := ec.ports[1].f

	for v := range in {
		fmt.Fprintln(out, toString(v))
	}
	return success
}

func feedchan(ec *evalCtx, args []Value) exitus {
	if len(args) > 0 {
		return argsError
	}
	in := ec.ports[0].f
	out := ec.ports[1].ch

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

func ratFn(ec *evalCtx, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	out := ec.ports[1].ch
	r, err := toRat(args[0])
	if err != nil {
		return newFailure(err.Error())
	}
	out <- r
	return success
}

// unpack takes any number of tables and output their list elements.
func unpack(ec *evalCtx, args []Value) exitus {
	if len(args) != 0 {
		return argsError
	}
	in := ec.ports[0].ch
	out := ec.ports[1].ch

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
func parseJSON(ec *evalCtx, args []Value) exitus {
	if len(args) > 0 {
		return argsError
	}
	in := ec.ports[0].f
	out := ec.ports[1].ch

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
func each(ec *evalCtx, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	if f, ok := args[0].(*closure); !ok {
		return argsError
	} else {
		in := ec.ports[0].ch
	in:
		for v := range in {
			su := f.Exec(ec.copy("closure of each"), []Value{v})
			// F.Exec will put exactly one stateUpdate on the channel
			e := (<-su).Exitus
			switch e.Sort {
			case Failure, Traceback, Return:
				return e
			case Success, Continue:
				// nop
			case Break:
				break in
			default:
				return newFailure(fmt.Sprintf("unknown exitusSort %v", e.Sort))
			}
		}
	}
	return success
}

func cd(ec *evalCtx, args []Value) exitus {
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
	if ec.store != nil {
		pwd, err := os.Getwd()
		// BUG(xiaq): Possible error of os.Getwd after cd-ing is ignored.
		if err == nil {
			ec.store.AddDir(pwd)
		}
	}
	return success
}

var storeNotConnected = newFailure("store not connected")

func visistedDirs(ec *evalCtx, args []Value) exitus {
	if ec.store == nil {
		return storeNotConnected
	}
	dirs, err := ec.store.ListDirs()
	if err != nil {
		return newFailure("store error: " + err.Error())
	}
	out := ec.ports[1].ch
	for _, dir := range dirs {
		table := newTable()
		table.Dict["path"] = str(dir.Path)
		table.Dict["score"] = str(fmt.Sprint(dir.Score))
		out <- table
	}
	return success
}

var noMatchingDir = newFailure("no matching directory")

func jumpDir(ec *evalCtx, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	if ec.store == nil {
		return storeNotConnected
	}
	dirs, err := ec.store.FindDirs(toString(args[0]))
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
	ec.store.AddDir(dir)
	return success
}

func source(ec *evalCtx, args []Value) exitus {
	if len(args) != 1 {
		return argsError
	}
	if fname, ok := args[0].(str); !ok {
		return argsError
	} else {
		ec.Source(string(fname))
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

func plus(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
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

func minus(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
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

func times(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
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

func divide(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
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

func eq(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
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

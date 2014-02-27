package eval

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/xiaq/elvish/parse"
)

type builtinFuncImpl func(*Evaluator, []Value, [2]*port) string

type builtinFunc struct {
	fn          builtinFuncImpl
	streamTypes [2]StreamType
}

type builtinSpecialImpl func(*Evaluator, []parse.Node, [2]*port) string

type builtinSpecial struct {
	fn          builtinSpecialImpl
	streamTypes [2]StreamType
}

var builtinFuncs = map[string]builtinFunc{
	"var":       builtinFunc{var_, [2]StreamType{unusedStream, unusedStream}},
	"set":       builtinFunc{set, [2]StreamType{unusedStream, unusedStream}},
	"fn":        builtinFunc{fn, [2]StreamType{unusedStream, unusedStream}},
	"put":       builtinFunc{put, [2]StreamType{unusedStream, chanStream}},
	"print":     builtinFunc{print, [2]StreamType{unusedStream}},
	"println":   builtinFunc{println, [2]StreamType{unusedStream}},
	"printchan": builtinFunc{printchan, [2]StreamType{chanStream, fdStream}},
	"feedchan":  builtinFunc{feedchan, [2]StreamType{fdStream, chanStream}},
	"cd":        builtinFunc{cd, [2]StreamType{unusedStream, unusedStream}},
	"+":         builtinFunc{plus, [2]StreamType{unusedStream, chanStream}},
	"-":         builtinFunc{minus, [2]StreamType{unusedStream, chanStream}},
	"*":         builtinFunc{times, [2]StreamType{unusedStream, chanStream}},
	"/":         builtinFunc{divide, [2]StreamType{unusedStream, chanStream}},
}

func doSet(ev *Evaluator, names []string, values []Value) string {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(names) != len(values) {
		return "arity mismatch"
	}

	for i, name := range names {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		ev.locals[name] = values[i]
	}

	return ""
}

func var_(ev *Evaluator, args []Value, ports [2]*port) string {
	var names []string
	var values []Value
	for i, nameVal := range args {
		name := nameVal.String(ev)
		if name == "=" {
			values = args[i+1:]
			break
		}
		if _, ok := ev.locals[name]; ok {
			return fmt.Sprintf("Variable %q already exists", name)
		}
		names = append(names, name)
	}

	for _, name := range names {
		ev.locals[name] = nil
	}
	if values != nil {
		return doSet(ev, names, values)
	}
	return ""
}

func set(ev *Evaluator, args []Value, ports [2]*port) string {
	var names []string
	var values []Value
	for i, nameVal := range args {
		name := nameVal.String(ev)
		if name == "=" {
			values = args[i+1:]
			break
		}
		if _, ok := ev.locals[name]; !ok {
			return fmt.Sprintf("Variable %q doesn't exists", name)
		}
		names = append(names, name)
	}

	if values == nil {
		return "missing equal sign"
	}
	return doSet(ev, names, values)
}

func fn(ev *Evaluator, args []Value, ports [2]*port) string {
	n := len(args)
	if n < 2 {
		return "args error"
	}
	closure, ok := args[n-1].(*Closure)
	if !ok {
		return "args error"
	}
	if n > 2 && len(closure.ArgNames) != 0 {
		return "can't define arg names list twice"
	}
	// BUG(xiaq): the fn builtin now modifies the closure in place, making it
	// possible to write:
	//
	// var f; set f = { }
	//
	// fn g a b $f // Changes arity of $f!
	for i := 1; i < n-1; i++ {
		closure.ArgNames = append(closure.ArgNames, args[i].String(ev))
	}
	// TODO(xiaq): should fn warn about redefinition of functions?
	ev.locals["fn-"+args[0].String(ev)] = closure
	return ""
}

func put(ev *Evaluator, args []Value, ports [2]*port) string {
	out := ports[1].ch
	for _, a := range args {
		out <- a
	}
	return ""
}

func print(ev *Evaluator, args []Value, ports [2]*port) string {
	out := ports[1].f
	for _, a := range args {
		fmt.Fprint(out, a.String(ev))
	}
	return ""
}

func println(ev *Evaluator, args []Value, ports [2]*port) string {
	args = append(args, NewString("\n"))
	return print(ev, args, ports)
}

func printchan(ev *Evaluator, args []Value, ports [2]*port) string {
	if len(args) > 0 {
		return "args error"
	}
	in := ports[0].ch
	out := ports[1].f

	for s := range in {
		fmt.Fprintln(out, s.String(ev))
	}
	return ""
}

func feedchan(ev *Evaluator, args []Value, ports [2]*port) string {
	if len(args) > 0 {
		return "args error"
	}
	in := ports[0].f
	out := ports[1].ch

	fmt.Println("WARNING: Only string input is supported at the moment.")

	bufferedIn := bufio.NewReader(in)
	// i := 0
	for {
		// fmt.Printf("[%v] ", i)
		line, err := bufferedIn.ReadString('\n')
		if err == io.EOF {
			return ""
		} else if err != nil {
			return err.Error()
		}
		out <- NewString(line[:len(line)-1])
		// i++
	}
}

func cd(ev *Evaluator, args []Value, ports [2]*port) string {
	var dir string
	if len(args) == 0 {
		dir = ""
	} else if len(args) == 1 {
		dir = args[0].String(ev)
	} else {
		return "args error"
	}
	err := os.Chdir(dir)
	if err != nil {
		return err.Error()
	}
	return ""
}

func toFloats(args []Value) (nums []float64, err error) {
	for _, a := range args {
		a, ok := a.(*String)
		if !ok {
			return nil, fmt.Errorf("must be string")
		}
		f, err := strconv.ParseFloat(string(*a), 64)
		if err != nil {
			return nil, err
		}
		nums = append(nums, f)
	}
	return
}

func plus(ev *Evaluator, args []Value, ports [2]*port) string {
	out := ports[1].ch
	nums, err := toFloats(args)
	if err != nil {
		return err.Error()
	}
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	out <- NewString(fmt.Sprintf("%g", sum))
	return ""
}

func minus(ev *Evaluator, args []Value, ports [2]*port) string {
	out := ports[1].ch
	if len(args) == 0 {
		return "not enough args"
	}
	nums, err := toFloats(args)
	if err != nil {
		return err.Error()
	}
	sum := nums[0]
	for _, f := range nums[1:] {
		sum -= f
	}
	out <- NewString(fmt.Sprintf("%g", sum))
	return ""
}

func times(ev *Evaluator, args []Value, ports [2]*port) string {
	out := ports[1].ch
	nums, err := toFloats(args)
	if err != nil {
		return err.Error()
	}
	prod := 1.0
	for _, f := range nums {
		prod *= f
	}
	out <- NewString(fmt.Sprintf("%g", prod))
	return ""
}

func divide(ev *Evaluator, args []Value, ports [2]*port) string {
	out := ports[1].ch
	if len(args) == 0 {
		return "not enough args"
	}
	nums, err := toFloats(args)
	if err != nil {
		return err.Error()
	}
	prod := nums[0]
	for _, f := range nums[1:] {
		prod /= f
	}
	out <- NewString(fmt.Sprintf("%g", prod))
	return ""
}

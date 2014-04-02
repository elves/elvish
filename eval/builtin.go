package eval

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"

	"github.com/xiaq/elvish/parse"
)

type builtinFuncImpl func(*Evaluator, []Value, [2]*port) string

type builtinFunc struct {
	fn          builtinFuncImpl
	streamTypes [2]StreamType
}

type builtinSpecialImpl func(*Evaluator, *formAnnotation, [2]*port) string
type builtinSpecialCheck func(*Checker, *parse.FormNode) interface{}

type builtinSpecial struct {
	fn          builtinSpecialImpl
	check       builtinSpecialCheck
	streamTypes [2]StreamType
}

var builtinFuncs = map[string]builtinFunc{
	"fn":        builtinFunc{fn, [2]StreamType{}},
	"put":       builtinFunc{put, [2]StreamType{0, chanStream}},
	"print":     builtinFunc{print, [2]StreamType{0, fdStream}},
	"println":   builtinFunc{println, [2]StreamType{0, fdStream}},
	"printchan": builtinFunc{printchan, [2]StreamType{chanStream, fdStream}},
	"feedchan":  builtinFunc{feedchan, [2]StreamType{fdStream, chanStream}},
	"cd":        builtinFunc{cd, [2]StreamType{}},
	"+":         builtinFunc{plus, [2]StreamType{0, chanStream}},
	"-":         builtinFunc{minus, [2]StreamType{0, chanStream}},
	"*":         builtinFunc{times, [2]StreamType{0, chanStream}},
	"/":         builtinFunc{divide, [2]StreamType{0, chanStream}},
}

var builtinSpecials map[string]builtinSpecial

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]builtinSpecial{
		"var": builtinSpecial{var_, checkVar, [2]StreamType{}},
		"set": builtinSpecial{set, checkSet, [2]StreamType{}},
	}
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
	ev.scope["fn-"+args[0].String(ev)] = valuePtr(closure)
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
		user, err := user.Current()
		if err == nil {
			dir = user.HomeDir
		}
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

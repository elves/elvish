package eval

// Builtin functions.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
)

type builtinFuncImpl func(*Evaluator, []Value) string

type builtinFunc struct {
	fn          builtinFuncImpl
	streamTypes [2]StreamType
}

var builtinFuncs = map[string]builtinFunc{
	"put":       builtinFunc{put, [2]StreamType{0, chanStream}},
	"typeof":    builtinFunc{typeof, [2]StreamType{0, chanStream}},
	"print":     builtinFunc{print, [2]StreamType{0, fdStream}},
	"println":   builtinFunc{println, [2]StreamType{0, fdStream}},
	"printchan": builtinFunc{printchan, [2]StreamType{chanStream, fdStream}},
	"feedchan":  builtinFunc{feedchan, [2]StreamType{fdStream, chanStream}},
	"unpack":    builtinFunc{unpack, [2]StreamType{0, chanStream}},
	"each":      builtinFunc{each, [2]StreamType{chanStream, hybridStream}},
	"cd":        builtinFunc{cd, [2]StreamType{}},
	"+":         builtinFunc{plus, [2]StreamType{0, chanStream}},
	"-":         builtinFunc{minus, [2]StreamType{0, chanStream}},
	"*":         builtinFunc{times, [2]StreamType{0, chanStream}},
	"/":         builtinFunc{divide, [2]StreamType{0, chanStream}},
}

func put(ev *Evaluator, args []Value) string {
	out := ev.ports[1].ch
	for _, a := range args {
		out <- a
	}
	return ""
}

func typeof(ev *Evaluator, args []Value) string {
	out := ev.ports[1].ch
	for _, a := range args {
		out <- NewString(a.Type().String())
	}
	return ""
}

func print(ev *Evaluator, args []Value) string {
	out := ev.ports[1].f
	for _, a := range args {
		fmt.Fprint(out, a.String())
	}
	return ""
}

func println(ev *Evaluator, args []Value) string {
	args = append(args, NewString("\n"))
	return print(ev, args)
}

func printchan(ev *Evaluator, args []Value) string {
	if len(args) > 0 {
		return "args error"
	}
	in := ev.ports[0].ch
	out := ev.ports[1].f

	for s := range in {
		fmt.Fprintln(out, s.String())
	}
	return ""
}

func feedchan(ev *Evaluator, args []Value) string {
	if len(args) > 0 {
		return "args error"
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
			return ""
		} else if err != nil {
			return err.Error()
		}
		out <- NewString(line[:len(line)-1])
		// i++
	}
}

// unpack takes any number of tables and output their list elements.
func unpack(ev *Evaluator, args []Value) string {
	out := ev.ports[1].ch
	for _, a := range args {
		if _, ok := a.(*Table); !ok {
			return "args error"
		}
	}
	for _, a := range args {
		a := a.(*Table)
		for _, e := range a.List {
			out <- e
		}
	}
	return ""
}

// each takes a single closure and applies it to all input values.
func each(ev *Evaluator, args []Value) string {
	if len(args) != 1 {
		return "args error"
	}
	if f, ok := args[0].(*Closure); !ok {
		return "args error"
	} else {
		in := ev.ports[0].ch
		for v := range in {
			su := ev.execClosure(f, []Value{v})
			<-su
		}
	}
	return ""
}

func cd(ev *Evaluator, args []Value) string {
	var dir string
	if len(args) == 0 {
		user, err := user.Current()
		if err == nil {
			dir = user.HomeDir
		}
	} else if len(args) == 1 {
		dir = args[0].String()
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

func plus(ev *Evaluator, args []Value) string {
	out := ev.ports[1].ch
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

func minus(ev *Evaluator, args []Value) string {
	out := ev.ports[1].ch
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

func times(ev *Evaluator, args []Value) string {
	out := ev.ports[1].ch
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

func divide(ev *Evaluator, args []Value) string {
	out := ev.ports[1].ch
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

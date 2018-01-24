package edit

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/xiaq/persistent/hash"
)

// For an overview of completion, see the comment in completers.go.
//
// When completing arguments of commands (as opposed to variable names, map
// indicies, etc.), the list of candidates often depends on the command in
// question; e.g. "ls <Tab>" and "apt <Tab>" should yield different results
// because the commands "ls" and "apt" accept different arguments. To reflec
// this, Elvish has a map of "argument completers", with the key being the name
// of the command, and the value being the argument completer itself, accessible
// to script as $edit:arg-completer. The one entry with an empty string as the
// key is the fallback completer, and is used when an argument completer for the
// current command has not been defined.
//
// When completing an argument, Elvish first finds out the name of the command
// (e.g. "ls" or "apt") can tries to evaluate its arguments. It then calls the
// suitable completer with the name of the command and the arguments. The
// arguments are in evaluated forms: e.g. if an argument is 'foo' (with quotes),
// the argument is its value foo, not a literal 'foo'. The last argument is what
// needs to be completed; if the user is starting a new argument, e.g. by typing
// "ls a " (without quotes), the last argument passed to the argument completer
// will be an empty string.
//
// The argument completer should then return a list of what can replace the last
// argument. The results are of type rawCandidate, which basically means that
// argument completers do not need to worry about quoting of candidates; the raw
// candidates will be "cooked" into actual candidates that appear in the
// interface, which includes quoting.
//
// There are a few finer points in this process:
//
// 1. If some of the arguments cannot be evaluated statically (for instance,
//    consider this: echo (uname)), it will be an empty string. There needs
//    probably be a better way to distinguish empty arguments and unknown
//    arguments, but normally there is not much argument completers can do for
//    unknown arguments.
//
// 2. The argument completer normally **should not** perform filtering. For
//    instance, if the user has typed "ls x", the argument completer for "ls"
//    should return **all** files, not just those whose names start with x. This
//    is to make it possible for user to specify a different matching algorithm
//    than the default prefix matching.
//
//    However, argument completers **should** look at the argument to decide
//    which **type** of candidates to generate. For instance, if the user has
//    typed "ls --x", the argument completer should generate all long options
//    for "ls", but not only those starting with "x".

var (
	// ErrCompleterMustBeFn is thrown if the user has put a non-function entry
	// in $edit:completer, and that entry needs to be used for completion.
	// TODO(xiaq): Detect the type violation when the user modifies
	// $edit:completer.
	ErrCompleterMustBeFn = errors.New("completer must be fn")
	// ErrCompleterArgMustBeString is thrown when a builtin argument completer
	// is called with non-string arguments.
	ErrCompleterArgMustBeString = errors.New("arguments to arg completers must be string")
	// ErrTooFewArguments is thrown when a builtin argument completer is called
	// with too few arguments.
	ErrTooFewArguments = errors.New("too few arguments")
)

var (
	argCompletersData = map[string]*builtinArgCompleter{
		"":     {"complete-filename", complFilename},
		"sudo": {"complete-sudo", complSudo},
	}
)

var _ = RegisterVariable("arg-completer", argCompleterVariable)

func argCompleterVariable() vartypes.Variable {
	m := types.EmptyMapInner
	for k, v := range argCompletersData {
		m = m.Assoc(types.String(k), v)
	}
	return vartypes.NewValidatedPtr(types.NewMap(m), vartypes.ShouldBeMap)
}

func (ed *Editor) argCompleter() types.Map {
	return ed.variables["arg-completer"].Get().(types.Map)
}

// completeArg calls the correct argument completers according to the command
// name. It is used by complArg and can also be useful when further dispatching
// based on command name is needed -- e.g. in the argument completer for "sudo".
func completeArg(words []string, ev *eval.Evaler, rawCands chan<- rawCandidate) error {
	logger.Printf("completing argument: %q", words)
	// XXX(xiaq): not the best way to get argCompleter.
	m := ev.Editor.(*Editor).argCompleter()
	var v types.Value
	index := words[0]
	if !m.HasKey(types.String(index)) {
		index = ""
	}
	v, err := m.Index(types.String(index))
	if err != nil {
		return err
	}
	fn, ok := v.(eval.Fn)
	if !ok {
		return ErrCompleterMustBeFn
	}
	return callArgCompleter(fn, ev, words, rawCands)
}

type builtinArgCompleter struct {
	name string
	impl func([]string, *eval.Evaler, chan<- rawCandidate) error
}

var _ eval.Fn = &builtinArgCompleter{}

func (bac *builtinArgCompleter) Kind() string {
	return "fn"
}

// Equal compares by identity.
func (bac *builtinArgCompleter) Equal(a interface{}) bool {
	return bac == a
}

func (bac *builtinArgCompleter) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(bac))
}

func (bac *builtinArgCompleter) Repr(int) string {
	return "$edit:" + bac.name + eval.FnSuffix
}

func (bac *builtinArgCompleter) Call(ec *eval.Frame, args []types.Value, opts map[string]types.Value) error {
	eval.TakeNoOpt(opts)
	words := make([]string, len(args))
	for i, arg := range args {
		s, ok := arg.(types.String)
		if !ok {
			throw(ErrCompleterArgMustBeString)
		}
		words[i] = string(s)
	}

	rawCands := make(chan rawCandidate)
	var err error
	go func() {
		defer close(rawCands)
		err = bac.impl(words, ec.Evaler, rawCands)
	}()

	output := ec.OutputChan()
	for rc := range rawCands {
		output <- rc
	}
	return err
}

func complFilename(words []string, ev *eval.Evaler, rawCands chan<- rawCandidate) error {
	if len(words) < 1 {
		return ErrTooFewArguments
	}
	return complFilenameInner(words[len(words)-1], false, rawCands)
}

func complSudo(words []string, ev *eval.Evaler, rawCands chan<- rawCandidate) error {
	if len(words) < 2 {
		return ErrTooFewArguments
	}
	if len(words) == 2 {
		return complFormHeadInner(words[1], ev, rawCands)
	}
	return completeArg(words[1:], ev, rawCands)
}

// callArgCompleter calls a Fn, assuming that it is an arg completer. It calls
// the Fn with specified arguments and closed input, and converts its output to
// candidate objects.
func callArgCompleter(fn eval.Fn,
	ev *eval.Evaler, words []string, rawCands chan<- rawCandidate) error {

	// Quick path for builtin arg completers.
	if builtin, ok := fn.(*builtinArgCompleter); ok {
		return builtin.impl(words, ev, rawCands)
	}

	args := make([]types.Value, len(words))
	for i, word := range words {
		args[i] = types.String(word)
	}

	ports := []*eval.Port{
		eval.DevNullClosedChan,
		{}, // Will be replaced when capturing output
		{File: os.Stderr},
	}

	valuesCb := func(ch <-chan types.Value) {
		for v := range ch {
			switch v := v.(type) {
			case rawCandidate:
				rawCands <- v
			case types.String:
				rawCands <- plainCandidate(v)
			default:
				logger.Printf("completer must output string or candidate")
			}
		}
	}

	bytesCb := func(r *os.File) {
		buffered := bufio.NewReader(r)
		for {
			line, err := buffered.ReadString('\n')
			if line != "" {
				rawCands <- plainCandidate(strings.TrimSuffix(line, "\n"))
			}
			if err != nil {
				if err != io.EOF {
					logger.Println("error on reading:", err)
				}
				break
			}
		}
	}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopFrame(ev, eval.NewInternalSource("[editor completer]"), ports)
	err := ec.PCaptureOutputInner(fn, args, eval.NoOpts, valuesCb, bytesCb)
	if err != nil {
		err = errors.New("completer error: " + err.Error())
	}

	return err
}

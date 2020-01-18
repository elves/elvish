package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/elves/elvish/pkg/eval/vals"
)

// Input and output.

//elvdoc:fn put
//
// ```elvish
// put $value...
// ```
//
// Takes arbitrary arguments and write them to the structured stdout.
//
// Examples:
//
// ```elvish-transcript
// ~> put a
// ▶ a
// ~> put lorem ipsum [a b] { ls }
// ▶ lorem
// ▶ ipsum
// ▶ [a b]
// ▶ <closure 0xc4202607e0>
// ```
//
// Etymology: Various languages, in particular
// [C](https://manpages.debian.org/stretch/manpages-dev/puts.3.en.html) and
// [Ruby](https://ruby-doc.org/core-2.2.2/IO.html#method-i-puts) as `puts`.

// TODO(xiaq): Document read-upto.

//elvdoc:fn print
//
// ```elvish
// print &sep=' ' $value...
// ```
//
// Like `echo`, just without the newline.
//
// @cf echo
//
// Etymology: Various languages, in particular
// [Perl](https://perldoc.perl.org/functions/print.html) and
// [zsh](http://zsh.sourceforge.net/Doc/Release/Shell-Builtin-Commands.html), whose
// `print`s do not print a trailing newline.

//elvdoc:fn echo
//
// ```elvish
// echo &sep=' ' $value...
// ```
//
// Print all arguments, joined by the `sep` option, and followed by a newline.
//
// Examples:
//
// ```elvish-transcript
// ~> echo Hello   elvish
// Hello elvish
// ~> echo "Hello   elvish"
// Hello   elvish
// ~> echo &sep=, lorem ipsum
// lorem,ipsum
// ```
//
// Notes: The `echo` builtin does not treat `-e` or `-n` specially. For instance,
// `echo -n` just prints `-n`. Use double-quoted strings to print special
// characters, and `print` to suppress the trailing newline.
//
// @cf print
//
// Etymology: Bourne sh.

//elvdoc:fn pprint
//
// ```elvish
// pprint $value...
// ```
//
// Pretty-print representations of Elvish values. Examples:
//
// ```elvish-transcript
// ~> pprint [foo bar]
// [
// foo
// bar
// ]
// ~> pprint [&k1=v1 &k2=v2]
// [
// &k2=
// v2
// &k1=
// v1
// ]
// ```
//
// The output format is subject to change.
//
// @cf repr

//elvdoc:fn repr
//
// ```elvish
// repr $value...
// ```
//
// Writes representation of `$value`s, separated by space and followed by a
// newline. Example:
//
// ```elvish-transcript
// ~> repr [foo 'lorem ipsum'] "aha\n"
// [foo 'lorem ipsum'] "aha\n"
// ```
//
// @cf pprint
//
// Etymology: [Python](https://docs.python.org/3/library/functions.html#repr).

//elvdoc:fn only-bytes
//
// ```elvish
// only-bytes
// ```
//
// Passes byte input to output, and discards value inputs.
//
// Example:
//
// ```elvish-transcript
// ~> { put value; echo bytes } | only-bytes
// bytes
// ```

//elvdoc:fn only-values
//
// ```elvish
// only-values
// ```
//
// Passes value input to output, and discards byte inputs.
//
// Example:
//
// ```elvish-transcript
// ~> { put value; echo bytes } | only-values
// ▶ value
// ```

//elvdoc:fn slurp
//
// ```elvish
// slurp
// ```
//
// Reads bytes input into a single string, and put this string on structured
// stdout.
//
// Example:
//
// ```elvish-transcript
// ~> echo "a\nb" | slurp
// ▶ "a\nb\n"
// ```
//
// Etymology: Perl, as
// [`File::Slurp`](http://search.cpan.org/~uri/File-Slurp-9999.19/lib/File/Slurp.pm).

// TODO(xiaq): Document from-lines.

//elvdoc:fn from-json
//
// ```elvish
// from-json
// ```
//
// Takes bytes stdin, parses it as JSON and puts the result on structured stdout.
// The input can contain multiple JSONs, which can, but do not have to, be
// separated with whitespaces.
//
// Examples:
//
// ```elvish-transcript
// ~> echo '"a"' | from-json
// ▶ a
// ~> echo '["lorem", "ipsum"]' | from-json
// ▶ [lorem ipsum]
// ~> echo '{"lorem": "ipsum"}' | from-json
// ▶ [&lorem=ipsum]
// ~> # multiple JSONs running together
// echo '"a""b"["x"]' | from-json
// ▶ a
// ▶ b
// ▶ [x]
// ~> # multiple JSONs separated by newlines
// echo '"a"
// {"k": "v"}' | from-json
// ▶ a
// ▶ [&k=v]
// ```
//
// @cf to-json

// TODO(xiaq): Document to-lines.

//elvdoc:fn to-json
//
// ```elvish
// to-json
// ```
//
// Takes structured stdin, convert it to JSON and puts the result on bytes stdout.
//
// ```elvish-transcript
// ~> put a | to-json
// "a"
// ~> put [lorem ipsum] | to-json
// ["lorem","ipsum"]
// ~> put [&lorem=ipsum] | to-json
// {"lorem":"ipsum"}
// ```
//
// @cf from-json

//elvdoc:fn fopen
//
// ```elvish
// fopen $filename
// ```
//
// Open a file. Currently, `fopen` only supports opening a file for reading. File
// must be closed with `fclose` explicitly. Example:
//
// ```elvish-transcript
// ~> cat a.txt
// This is
// a file.
// ~> f = (fopen a.txt)
// ~> cat < $f
// This is
// a file.
// ~> fclose $f
// ```
//
// @cf fclose

//elvdoc:fn fclose
//
// ```elvish
// fclose $file
// ```
//
// Close a file opened with `fopen`.
//
// @cf fopen

//elvdoc:fn pipe
//
// ```elvish
// pipe
// ```
//
// Create a new Unix pipe that can be used in redirections.
//
// A pipe contains both the read FD and the write FD. When redirecting command
// input to a pipe with `<`, the read FD is used. When redirecting command output
// to a pipe with `>`, the write FD is used. It is not supported to redirect both
// input and output with `<>` to a pipe.
//
// Pipes have an OS-dependent buffer, so writing to a pipe without an active reader
// does not necessarily block. Pipes **must** be explicitly closed with `prclose`
// and `pwclose`.
//
// Putting values into pipes will cause those values to be discarded.
//
// Examples (assuming the pipe has a large enough buffer):
//
// ```elvish-transcript
// ~> p = (pipe)
// ~> echo 'lorem ipsum' > $p
// ~> head -n1 < $p
// lorem ipsum
// ~> put 'lorem ipsum' > $p
// ~> head -n1 < $p
// # blocks
// # $p should be closed with prclose and pwclose afterwards
// ```
//
// @cf prclose pwclose

//elvdoc:fn prclose
//
// ```elvish
// prclose $pipe
// ```
//
// Close the read end of a pipe.
//
// @cf pwclose pipe

//elvdoc:fn pwclose
//
// ```elvish
// pwclose $pipe
// ```
//
// Close the write end of a pipe.
//
// @cf prclose pipe

func init() {
	addBuiltinFns(map[string]interface{}{
		// Value output
		"put": put,

		// Bytes input
		"read-upto": readUpto,

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

func readUpto(fm *Frame, last string) (string, error) {
	if len(last) != 1 {
		return "", ErrArgs
	}
	in := fm.InputFile()
	var buf []byte
	for {
		var b [1]byte
		_, err := in.Read(b[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		buf = append(buf, b[0])
		if b[0] == last[0] {
			break
		}
	}
	return string(buf), nil
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

func fopen(name string) (vals.File, error) {
	// TODO support opening files for writing etc as well.
	return os.Open(name)
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

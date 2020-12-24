package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/strutil"
)

// Input and output.

func init() {
	addBuiltinFns(map[string]interface{}{
		// Value output
		"put": put,

		// Bytes input
		"read-upto": readUpto,
		"read-line": readLine,

		// Bytes output
		"print":  print,
		"printf": printf,
		"echo":   echo,
		"pprint": pprint,
		"repr":   repr,
		"show":   show,

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

func put(fm *Frame, args ...interface{}) {
	out := fm.OutputChan()
	for _, a := range args {
		out <- a
	}
}

//elvdoc:fn read-upto
//
// ```elvish
// read-upto $delim
// ```
//
// Reads byte input until `$delim` or end-of-file is encountered, and outputs
// the part of the input read as a string value. The output contains the
// trailing `$delim`, unless `read-upto` terminated at end-of-file.
//
// The `$delim` argument must be a single rune in the ASCII range.
//
// Examples:
//
// ```elvish-transcript
// ~> echo "a,b,c" | read-upto ","
// ▶ 'a,'
// ~> echo "foo\nbar" | read-upto "\n"
// ▶ "foo\n"
// ~> echo "a.elv\x00b.elv" | read-upto "\x00"
// ▶ "a.elv\x00"
// ~> print "foobar" | read-upto "\n"
// ▶ foobar
// ```

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

//elvdoc:fn read-line
//
// ```elvish
// read-line
// ```
//
// Reads a single line from byte input, and writes the line to the value output,
// stripping the line ending. A line can end with `"\r\n"`, `"\n"`, or end of
// file. Examples:
//
// ```elvish-transcript
// ~> print line | read-line
// ▶ line
// ~> print "line\n" | read-line
// ▶ line
// ~> print "line\r\n" | read-line
// ▶ line
// ~> print "line-with-extra-cr\r\r\n" | read-line
// ▶ "line-with-extra-cr\r"
// ```

func readLine(fm *Frame) (string, error) {
	s, err := readUpto(fm, "\n")
	if err != nil {
		return "", err
	}
	return strutil.ChopLineEnding(s), nil
}

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

type printOpts struct{ Sep string }

func (o *printOpts) SetDefaultOptions() { o.Sep = " " }

func print(fm *Frame, opts printOpts, args ...interface{}) {
	out := fm.OutputFile()
	for i, arg := range args {
		if i > 0 {
			out.WriteString(opts.Sep)
		}
		out.WriteString(vals.ToString(arg))
	}
}

//elvdoc:fn printf
//
// ```elvish
// printf $fmt $value...
// ```
//
// Like [`print`](#print) but uses a formatting template to control how the
// values appear in the output. An implicit newline is not added. You must
// include a newline in the format (e.g., `printf "%s\n" yes`) or a string
// value if a newline is desired in the output. This command writes its output
// to the byte stream.
//
// See Go's [`fmt`](https://golang.org/pkg/fmt/#hdr-Printing) package for
// details about the formatting verbs and the various flags that modify the
// default behavior; e.g., left versus right justification. We explicitly
// handle only strings and float64 data types. Printing other types is likely
// to result in unexpected output (but not an exception). You can use the
// numeric formatting verbs (e.g., `%d`, `%f`) with either a float64 or a
// string that can be converted to a number. If you use a float64 value to
// the `%s` verb it will first be converted to a string using the rules for
// [numbers](language.html#number]).
//
// Examples:
//
// ```elvish-transcript
// ~> printf "%10s %.2f\n" Pi $math:pi
//         Pi 3.14
// ~> printf "%-10s %.2f %s\n" Pi $math:pi $math:pi
// Pi         3.14 3.141592653589793
// ~> printf "%d\n" 0b11100111
// 231
// ~> printf "%08b\n" 231
// 11100111
// ```
//
// **Note**: A couple of Go's [formatting
// verbs](https://golang.org/pkg/fmt/#hdr-Printing), such as `%b` and `%x`,
// behave differently based on the associated value type. In Elvish the
// type is always an integer; specifically, a Go `int`; regardless of whether
// the Elvish value is a string or float64. That is, the number is coerced to
// an int if possible else an exception is raised.
//
// **Note**: Do not use the `%p` or `%q` formatting verbs at this time. The `%p`
// verb will never have any meaning. The `%q` formatting verb will eventually
// behave as if you had printed the result of `repr` on the value. Similarly,
// only scalars such as strings and numbers are currently supported. Passing
// an Elvish list, map, exception, or other complex type will produce
// unexpected output. This will be improved in the future.
//
// **Note**: This is loosely based on the [POSIX `printf`
// command](https://pubs.opengroup.org/onlinepubs/007908799/xcu/printf.html).
// There are three notable differences. The first is that the formatting verbs
// and behavior use Go's [`fmt`](https://golang.org/pkg/fmt/) package and not
// the POSIX specification. The two have many similarities but are not
// identical. The second is that the number of values must match the number of
// formatting verbs. Excess values do not result in the format being evaluated
// additional times until all values are consumed as happens with the POSIX
// command of the same name. You must explicitly split the value list and
// invoke this builtin for each block of values. The third is that the Elvish
// `printf` does not recognize backslash sequences such as `\n`. You must use
// [double-quoted Elvish strings](language.html#double-quoted-string) if you
// want such backslash sequences to be recognized -- just as you would with a
// string passed to any other Elvish builtin.
//
// @cf print echo pprint repr

type formatterString struct {
	str string
}
type formatterFloat64 struct {
	num float64
}

func printf(fm *Frame, template string, args ...interface{}) {
	out := fm.OutputFile()
	aliasedArgs := make([]interface{}, len(args))
	// TODO: Implement support for `%q` but using Elvish's rules for quoting
	// strings such as produced by `repr '\a"b'"\033'"` rather than Go's
	// default `%q` behavior.
	//
	// TODO: Possibly special case other Elvish types such as lists, maps and
	// exceptions.
	for i, a := range args {
		switch a := a.(type) {
		case string:
			aliasedArgs[i] = formatterString{a}
		case float64:
			aliasedArgs[i] = formatterFloat64{a}
		default:
			// This will pretty much only work if the `%v` fmt verb is used or
			// the argument type matches the formatting verg (e.g., passing
			// `$true` to the `%t` verb).
			aliasedArgs[i] = a
		}
	}
	s := fmt.Sprintf(template, aliasedArgs...)
	out.WriteString(s)
}

// Convert a formatting verb state to a string that is logically equivalent to
// the original formatting verb.
func buildFmtVerb(state fmt.State, v rune) string {
	var flags = []rune{'%'}
	for _, f := range "+-# 0" {
		if state.Flag(int(f)) {
			flags = append(flags, f)
		}
	}
	s := string(flags)
	if w, ok := state.Width(); ok {
		s = fmt.Sprintf("%s%d", s, w)
	}
	if p, ok := state.Precision(); ok {
		s = fmt.Sprintf("%s.%d", s, p)
	}
	return s + string(v)
}

func stringAsFloat(state fmt.State, r rune, s string) {
	var n float64
	if err := vals.ScanToGo(s, &n); err != nil {
		fmt.Fprintf(state, "%%!f(%s)", err.Error())
		return
	}
	verb := buildFmtVerb(state, r)
	fmt.Fprintf(state, verb, n)
}

func stringAsInt(state fmt.State, r rune, s string) {
	var n int
	if err := vals.ScanToGo(s, &n); err != nil {
		fmt.Fprintf(state, "%%!%c(%s)", r, err.Error())
		return
	}
	verb := buildFmtVerb(state, r)
	fmt.Fprintf(state, verb, n)
}

// Format an Elvish float64 according to the provided fmt verb.
func (v formatterFloat64) Format(state fmt.State, r rune) {
	switch r {
	case 's', 'v':
		verb := buildFmtVerb(state, r)
		fmt.Fprintf(state, verb, vals.ToString(v.num))
	case 'e', 'E', 'f', 'F', 'g', 'G':
		verb := buildFmtVerb(state, r)
		fmt.Fprintf(state, verb, v.num)
	case 'b', 'd', 'o', 'O', 'x', 'X', 'U':
		i := int(v.num)
		if float64(i) != v.num {
			fmt.Fprintf(state, "%%!d(must be an integer)")
			return
		}
		verb := buildFmtVerb(state, r)
		fmt.Fprintf(state, verb, i)
	default:
		fmt.Fprintf(state, "%%!%c(unsupported formatting verb)", r)
	}
}

func (v formatterString) Format(state fmt.State, r rune) {
	switch r {
	case 's', 'v':
		verb := buildFmtVerb(state, r)
		fmt.Fprintf(state, verb, v.str)
	case 'e', 'E', 'f', 'F', 'g', 'G':
		stringAsFloat(state, r, v.str)
	case 'b', 'd', 'o', 'O', 'x', 'X', 'U':
		stringAsInt(state, r, v.str)
	default:
		fmt.Fprintf(state, "%%!%c(unsupported formatting verb)", r)
	}
}

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

func echo(fm *Frame, opts printOpts, args ...interface{}) {
	print(fm, opts, args...)
	fm.OutputFile().WriteString("\n")
}

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

func pprint(fm *Frame, args ...interface{}) {
	out := fm.OutputFile()
	for _, arg := range args {
		out.WriteString(vals.Repr(arg, 0))
		out.WriteString("\n")
	}
}

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

func repr(fm *Frame, args ...interface{}) {
	out := fm.OutputFile()
	for i, arg := range args {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(vals.Repr(arg, vals.NoPretty))
	}
	out.WriteString("\n")
}

//elvdoc:fn show
//
// ```elvish
// show $e
// ```
//
// Shows the value to the output, which is assumed to be a VT-100-compatible
// terminal.
//
// Currently, the only type of value that can be showed is exceptions, but this
// will likely expand in future.
//
// Example:
//
// ```elvish-transcript
// ~> e = ?(fail lorem-ipsum)
// ~> show $e
// Exception: lorem-ipsum
// [tty 3], line 1: e = ?(fail lorem-ipsum)
// ```

func show(fm *Frame, v diag.Shower) {
	fm.OutputFile().WriteString(v.Show(""))
	fm.OutputFile().WriteString("\n")
}

const bytesReadBufferSize = 512

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

func slurp(fm *Frame) (string, error) {
	b, err := ioutil.ReadAll(fm.InputFile())
	return string(b), err
}

//elvdoc:fn from-lines
//
// ```elvish
// from-lines
// ```
//
// Splits byte input into lines, and writes them to the value output. Value
// input is ignored.
//
// ```elvish-transcript
// ~> { echo a; echo b } | from-lines
// ▶ a
// ▶ b
// ~> { echo a; put b } | from-lines
// ▶ a
// ```
//
// @cf to-lines

func fromLines(fm *Frame) {
	linesToChan(fm.InputFile(), fm.OutputChan())
}

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

func fromJSON(fm *Frame) error {
	in := fm.InputFile()
	out := fm.OutputChan()

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

// Converts a interface{} that results from json.Unmarshal to an Elvish value.
func fromJSONInterface(v interface{}) (interface{}, error) {
	switch v := v.(type) {
	case nil, bool, string:
		return v, nil
	case float64:
		return v, nil
	case []interface{}:
		vec := vals.EmptyList
		for _, elem := range v {
			converted, err := fromJSONInterface(elem)
			if err != nil {
				return nil, err
			}
			vec = vec.Cons(converted)
		}
		return vec, nil
	case map[string]interface{}:
		m := vals.EmptyMap
		for key, val := range v {
			convertedVal, err := fromJSONInterface(val)
			if err != nil {
				return nil, err
			}
			m = m.Assoc(key, convertedVal)
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unexpected json type: %T", v)
	}
}

//elvdoc:fn to-lines
//
// ```elvish
// to-lines $input?
// ```
//
// Writes each value input to a separate line in the byte output. Byte input is
// ignored.
//
// ```elvish-transcript
// ~> put a b | to-lines
// a
// b
// ~> to-lines [a b]
// a
// b
// ~> { put a; echo b } | to-lines
// b
// a
// ```
//
// @cf from-lines

func toLines(fm *Frame, inputs Inputs) {
	out := fm.OutputFile()

	inputs(func(v interface{}) {
		fmt.Fprintln(out, vals.ToString(v))
	})
}

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

func fopen(name string) (vals.File, error) {
	// TODO support opening files for writing etc as well.
	return os.Open(name)
}

//elvdoc:fn fclose
//
// ```elvish
// fclose $file
// ```
//
// Close a file opened with `fopen`.
//
// @cf fopen

func fclose(f vals.File) error {
	return f.Close()
}

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

func pipe() (vals.Pipe, error) {
	r, w, err := os.Pipe()
	return vals.NewPipe(r, w), err
}

//elvdoc:fn prclose
//
// ```elvish
// prclose $pipe
// ```
//
// Close the read end of a pipe.
//
// @cf pwclose pipe

func prclose(p vals.Pipe) error {
	return p.ReadEnd.Close()
}

//elvdoc:fn pwclose
//
// ```elvish
// pwclose $pipe
// ```
//
// Close the write end of a pipe.
//
// @cf prclose pipe

func pwclose(p vals.Pipe) error {
	return p.WriteEnd.Close()
}

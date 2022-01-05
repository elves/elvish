package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/strutil"
)

// Input and output.

func init() {
	addBuiltinFns(map[string]interface{}{
		// Value output
		"put":    put,
		"repeat": repeat,

		// Bytes input
		"read-upto": readUpto,
		"read-line": readLine,

		// Bytes output
		"print":  print,
		"echo":   echo,
		"pprint": pprint,
		"repr":   repr,
		"show":   show,
		"printf": printf,

		// Only bytes or values
		//
		// These are now implemented as commands forwarding one part of input to
		// output and discarding the other. A future optimization the evaler can
		// do is to connect the relevant parts directly together without any
		// kind of forwarding.
		"only-bytes":  onlyBytes,
		"only-values": onlyValues,

		// Bytes to value
		"slurp":           slurp,
		"from-lines":      fromLines,
		"from-json":       fromJSON,
		"from-terminated": fromTerminated,

		// Value to bytes
		"to-lines":      toLines,
		"to-json":       toJSON,
		"to-terminated": toTerminated,
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
// **Note**: It is almost never necessary to use `put (...)` - just write the
// `...` part. For example, `put (eq a b)` is the equivalent to just `eq a b`.
//
// Etymology: Various languages, in particular
// [C](https://manpages.debian.org/stretch/manpages-dev/puts.3.en.html) and
// [Ruby](https://ruby-doc.org/core-2.2.2/IO.html#method-i-puts) as `puts`.

func put(fm *Frame, args ...interface{}) error {
	out := fm.ValueOutput()
	for _, a := range args {
		err := out.Put(a)
		if err != nil {
			return err
		}
	}
	return nil
}

//elvdoc:fn repeat
//
// ```elvish
// repeat $n $value
// ```
//
// Output `$value` for `$n` times. Example:
//
// ```elvish-transcript
// ~> repeat 0 lorem
// ~> repeat 4 NAN
// ▶ NAN
// ▶ NAN
// ▶ NAN
// ▶ NAN
// ```
//
// Etymology: [Clojure](https://clojuredocs.org/clojure.core/repeat).

func repeat(fm *Frame, n int, v interface{}) error {
	out := fm.ValueOutput()
	for i := 0; i < n; i++ {
		err := out.Put(v)
		if err != nil {
			return err
		}
	}
	return nil
}

//elvdoc:fn read-upto
//
// ```elvish
// read-upto $terminator
// ```
//
// Reads byte input until `$terminator` or end-of-file is encountered. It outputs the part of the
// input read as a string value. The output contains the trailing `$terminator`, unless `read-upto`
// terminated at end-of-file.
//
// The `$terminator` must be a single ASCII character such as `"\x00"` (NUL).
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

func readUpto(fm *Frame, terminator string) (string, error) {
	if err := checkTerminator(terminator); err != nil {
		return "", err
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
		if b[0] == terminator[0] {
			break
		}
	}
	return string(buf), nil
}

func checkTerminator(s string) error {
	if len(s) != 1 || s[0] > 127 {
		return errs.BadValue{What: "terminator",
			Valid: "a single ASCII character", Actual: parse.Quote(s)}
	}
	return nil
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

func print(fm *Frame, opts printOpts, args ...interface{}) error {
	out := fm.ByteOutput()
	for i, arg := range args {
		if i > 0 {
			_, err := out.WriteString(opts.Sep)
			if err != nil {
				return err
			}
		}
		_, err := out.WriteString(vals.ToString(arg))
		if err != nil {
			return err
		}
	}
	return nil
}

//elvdoc:fn printf
//
// ```elvish
// printf $template $value...
// ```
//
// Prints values to the byte stream according to a template. If you need to inject the output into
// the value stream use this pattern: `printf .... | slurp`. That ensures that any newlines in the
// output of `printf` do not cause its output to be broken into multiple values, thus eliminating
// the newlines, which will occur if you do `put (printf ....)`.
//
// Like [`print`](#print), this command does not add an implicit newline; include an explicit `"\n"`
// in the formatting template instead. For example, `printf "%.1f\n" (/ 10.0 3)`.
//
// See Go's [`fmt`](https://golang.org/pkg/fmt/#hdr-Printing) package for
// details about the formatting verbs and the various flags that modify the
// default behavior, such as padding and justification.
//
// Unlike Go, each formatting verb has a single associated internal type, and
// accepts any argument that can reasonably be converted to that type:
//
// - The verbs `%s`, `%q` and `%v` convert the corresponding argument to a
//   string in different ways:
//
//     - `%s` uses [to-string](#to-string) to convert a value to string.
//
//     - `%q` uses [repr](#repr) to convert a value to string.
//
//     - `%v` is equivalent to `%s`, and `%#v` is equivalent to `%q`.
//
// - The verb `%t` first convert the corresponding argument to a boolean using
//   [bool](#bool), and then uses its Go counterpart to format the boolean.
//
// - The verbs `%b`, `%c`, `%d`, `%o`, `%O`, `%x`, `%X` and `%U` first convert
//   the corresponding argument to an integer using an internal algorithm, and
//   use their Go counterparts to format the integer.
//
// - The verbs `%e`, `%E`, `%f`, `%F`, `%g` and `%G` first convert the
//   corresponding argument to a floating-point number using
//   [float64](#float64), and then use their Go counterparts to format the
//   number.
//
// The special verb `%%` prints a literal `%` and consumes no argument.
//
// Verbs not documented above are not supported.
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
// ~> printf "list is: %q\n" [foo bar 'foo bar']
// list is: [foo bar 'foo bar']
// ```
//
// **Note**: Compared to the [POSIX `printf`
// command](https://pubs.opengroup.org/onlinepubs/007908799/xcu/printf.html)
// found in other shells, there are 3 key differences:
//
// - The behavior of the formatting verbs are based on Go's
//   [`fmt`](https://golang.org/pkg/fmt/) package instead of the POSIX
//   specification.
//
// - The number of arguments after the formatting template must match the number
//   of formatting verbs. The POSIX command will repeat the template string to
//   consume excess values; this command does not have that behavior.
//
// - This command does not interpret escape sequences such as `\n`; just use
//   [double-quoted strings](language.html#double-quoted-string).
//
// @cf print echo pprint repr

func printf(fm *Frame, template string, args ...interface{}) error {
	wrappedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		wrappedArgs[i] = formatter{arg}
	}

	_, err := fmt.Fprintf(fm.ByteOutput(), template, wrappedArgs...)
	return err
}

type formatter struct {
	wrapped interface{}
}

func (f formatter) Format(state fmt.State, r rune) {
	wrapped := f.wrapped
	switch r {
	case 's':
		writeFmt(state, 's', vals.ToString(wrapped))
	case 'q':
		// TODO: Support using the precision flag to specify indentation.
		writeFmt(state, 's', vals.ReprPlain(wrapped))
	case 'v':
		var s string
		if state.Flag('#') {
			s = vals.ReprPlain(wrapped)
		} else {
			s = vals.ToString(wrapped)
		}
		writeFmt(state, 's', s)
	case 't':
		writeFmt(state, 't', vals.Bool(wrapped))
	case 'b', 'c', 'd', 'o', 'O', 'x', 'X', 'U':
		var i int
		if err := vals.ScanToGo(wrapped, &i); err != nil {
			fmt.Fprintf(state, "%%!%c(%s)", r, err.Error())
			return
		}
		writeFmt(state, r, i)
	case 'e', 'E', 'f', 'F', 'g', 'G':
		var f float64
		if err := vals.ScanToGo(wrapped, &f); err != nil {
			fmt.Fprintf(state, "%%!%c(%s)", r, err.Error())
			return
		}
		writeFmt(state, r, f)
	default:
		fmt.Fprintf(state, "%%!%c(unsupported formatting verb)", r)
	}
}

// Writes to State using the flag it stores, but with a potentially different
// verb and value.
func writeFmt(state fmt.State, v rune, val interface{}) {
	// Reconstruct the verb string.
	var sb strings.Builder
	sb.WriteRune('%')
	for _, f := range "+-# 0" {
		if state.Flag(int(f)) {
			sb.WriteRune(f)
		}
	}
	if w, ok := state.Width(); ok {
		sb.WriteString(strconv.Itoa(w))
	}
	if p, ok := state.Precision(); ok {
		sb.WriteRune('.')
		sb.WriteString(strconv.Itoa(p))
	}
	sb.WriteRune(v)

	fmt.Fprintf(state, sb.String(), val)
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

func echo(fm *Frame, opts printOpts, args ...interface{}) error {
	err := print(fm, opts, args...)
	if err != nil {
		return err
	}
	_, err = fm.ByteOutput().WriteString("\n")
	return err
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

func pprint(fm *Frame, args ...interface{}) error {
	out := fm.ByteOutput()
	for _, arg := range args {
		_, err := out.WriteString(vals.Repr(arg, 0))
		if err != nil {
			return err
		}
		_, err = out.WriteString("\n")
		if err != nil {
			return err
		}
	}
	return nil
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

func repr(fm *Frame, args ...interface{}) error {
	out := fm.ByteOutput()
	for i, arg := range args {
		if i > 0 {
			_, err := out.WriteString(" ")
			if err != nil {
				return err
			}
		}
		_, err := out.WriteString(vals.ReprPlain(arg))
		if err != nil {
			return err
		}
	}
	_, err := out.WriteString("\n")
	return err
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
// ~> var e = ?(fail lorem-ipsum)
// ~> show $e
// Exception: lorem-ipsum
// [tty 3], line 1: var e = ?(fail lorem-ipsum)
// ```

func show(fm *Frame, v diag.Shower) error {
	out := fm.ByteOutput()
	_, err := out.WriteString(v.Show(""))
	if err != nil {
		return err
	}
	_, err = out.WriteString("\n")
	return err
}

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

	_, err := io.Copy(fm.ByteOutput(), fm.InputFile())
	return err
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
	// Discard bytes in a goroutine.
	bytesDone := make(chan struct{})
	go func() {
		// Ignore the error
		_, _ = io.Copy(blackholeWriter{}, fm.InputFile())
		close(bytesDone)
	}()
	// Wait for the goroutine to finish before returning.
	defer func() { <-bytesDone }()

	// Forward values.
	out := fm.ValueOutput()
	for v := range fm.InputChan() {
		err := out.Put(v)
		if err != nil {
			return err
		}
	}
	return nil
}

type blackholeWriter struct{}

func (blackholeWriter) Write(p []byte) (int, error) { return len(p), nil }

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
	b, err := io.ReadAll(fm.InputFile())
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
// @cf from-terminated read-upto to-lines

func fromLines(fm *Frame) error {
	filein := bufio.NewReader(fm.InputFile())
	out := fm.ValueOutput()
	for {
		line, err := filein.ReadString('\n')
		if line != "" {
			err := out.Put(strutil.ChopLineEnding(line))
			if err != nil {
				return err
			}
		}
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}

//elvdoc:fn from-json
//
// ```elvish
// from-json
// ```
//
// Takes bytes stdin, parses it as JSON and puts the result on structured stdout.
// The input can contain multiple JSONs, and whitespace between them are ignored.
//
// Note that JSON's only number type corresponds to Elvish's floating-point
// number type, and is always considered [inexact](language.html#exactness).
// It may be necessary to coerce JSON numbers to exact numbers using
// [exact-num](#exact-num).
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
	out := fm.ValueOutput()

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
		err = out.Put(converted)
		if err != nil {
			return err
		}
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
			vec = vec.Conj(converted)
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

//elvdoc:fn from-terminated
//
// ```elvish
// from-terminated $terminator
// ```
//
// Splits byte input into lines at each `$terminator` character, and writes
// them to the value output. If the byte input ends with `$terminator`, it is
// dropped. Value input is ignored.
//
// The `$terminator` must be a single ASCII character such as `"\x00"` (NUL).
//
// ```elvish-transcript
// ~> { echo a; echo b } | from-terminated "\x00"
// ▶ "a\nb\n"
// ~> print "a\x00b" | from-terminated "\x00"
// ▶ a
// ▶ b
// ~> print "a\x00b\x00" | from-terminated "\x00"
// ▶ a
// ▶ b
// ```
//
// @cf from-lines read-upto to-terminated

func fromTerminated(fm *Frame, terminator string) error {
	if err := checkTerminator(terminator); err != nil {
		return err
	}

	filein := bufio.NewReader(fm.InputFile())
	out := fm.ValueOutput()
	for {
		line, err := filein.ReadString(terminator[0])
		if line != "" {
			err := out.Put(strutil.ChopTerminator(line, terminator[0]))
			if err != nil {
				return err
			}
		}
		if err != nil {
			if err != io.EOF {
				logger.Println("error on reading:", err)
				return err
			}
			return nil
		}
	}
}

//elvdoc:fn to-lines
//
// ```elvish
// to-lines $inputs?
// ```
//
// Writes each [value input](#value-inputs) to a separate line in the byte
// output. Byte input is ignored.
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
// @cf from-lines to-terminated

func toLines(fm *Frame, inputs Inputs) error {
	out := fm.ByteOutput()
	var errOut error

	inputs(func(v interface{}) {
		if errOut != nil {
			return
		}
		// TODO: Don't ignore the error.
		_, errOut = fmt.Fprintln(out, vals.ToString(v))
	})
	return errOut
}

//elvdoc:fn to-terminated
//
// ```elvish
// to-terminated $terminator $inputs?
// ```
//
// Writes each [value input](#value-inputs) to the byte output with the
// specified terminator character. Byte input is ignored. This behavior is
// useful, for example, when feeding output into a program that accepts NUL
// terminated lines to avoid ambiguities if the values contains newline
// characters.
//
// The `$terminator` must be a single ASCII character such as `"\x00"` (NUL).
//
// ```elvish-transcript
// ~> put a b | to-terminated "\x00" | slurp
// ▶ "a\x00b\x00"
// ~> to-terminated "\x00" [a b] | slurp
// ▶ "a\x00b\x00"
// ```
//
// @cf from-terminated to-lines

func toTerminated(fm *Frame, terminator string, inputs Inputs) error {
	if err := checkTerminator(terminator); err != nil {
		return err
	}

	out := fm.ByteOutput()
	var errOut error
	inputs(func(v interface{}) {
		if errOut != nil {
			return
		}
		_, errOut = fmt.Fprint(out, vals.ToString(v), terminator)
	})
	return errOut
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
	encoder := json.NewEncoder(fm.ByteOutput())

	var errEncode error
	inputs(func(v interface{}) {
		if errEncode != nil {
			return
		}
		errEncode = encoder.Encode(v)
	})
	return errEncode
}

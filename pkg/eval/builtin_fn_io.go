package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
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
	addBuiltinFns(map[string]any{
		// Value output
		"put":    put,
		"repeat": repeat,

		// Bytes input
		"read-bytes": readBytes,
		"read-upto":  readUpto,
		"read-line":  readLine,

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

func put(fm *Frame, args ...any) error {
	out := fm.ValueOutput()
	for _, a := range args {
		err := out.Put(a)
		if err != nil {
			return err
		}
	}
	return nil
}

func repeat(fm *Frame, n int, v any) error {
	out := fm.ValueOutput()
	for i := 0; i < n; i++ {
		err := out.Put(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func readBytes(fm *Frame, max int) (string, error) {
	in := fm.InputFile()
	buf := make([]byte, max)
	read := 0
	for read < max {
		n, err := in.Read(buf[read:])
		read += n
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
	}
	return string(buf[:read]), nil
}

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

func readLine(fm *Frame) (string, error) {
	s, err := readUpto(fm, "\n")
	if err != nil {
		return "", err
	}
	return strutil.ChopLineEnding(s), nil
}

type printOpts struct{ Sep string }

func (o *printOpts) SetDefaultOptions() { o.Sep = " " }

func print(fm *Frame, opts printOpts, args ...any) error {
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

func printf(fm *Frame, template string, args ...any) error {
	wrappedArgs := make([]any, len(args))
	for i, arg := range args {
		wrappedArgs[i] = formatter{arg}
	}

	_, err := fmt.Fprintf(fm.ByteOutput(), template, wrappedArgs...)
	return err
}

type formatter struct {
	wrapped any
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
func writeFmt(state fmt.State, v rune, val any) {
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

func echo(fm *Frame, opts printOpts, args ...any) error {
	err := print(fm, opts, args...)
	if err != nil {
		return err
	}
	_, err = fm.ByteOutput().WriteString("\n")
	return err
}

func pprint(fm *Frame, args ...any) error {
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

func repr(fm *Frame, args ...any) error {
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

func show(fm *Frame, v diag.Shower) error {
	out := fm.ByteOutput()
	_, err := out.WriteString(v.Show(""))
	if err != nil {
		return err
	}
	_, err = out.WriteString("\n")
	return err
}

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

func slurp(fm *Frame) (string, error) {
	b, err := io.ReadAll(fm.InputFile())
	return string(b), err
}

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

func fromJSON(fm *Frame) error {
	in := fm.InputFile()
	out := fm.ValueOutput()

	dec := json.NewDecoder(in)
	// See comments below about using json.Number.
	dec.UseNumber()
	for {
		var v any
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
func fromJSONInterface(v any) (any, error) {
	switch v := v.(type) {
	case nil, bool, string:
		return v, nil
	case json.Number:
		// The JSON syntax doesn't restrict the precision of numbers. Since
		// we called json.Decoder.UseNumber, it preserves the full number
		// literal, and we can try parsing it as a big int.
		if z, ok := new(big.Int).SetString(v.String(), 0); ok {
			// Also normalize to int if the value fits.
			return vals.NormalizeBigInt(z), nil
		}
		// Parse as float64 instead. This can error if the number is not an
		// integer and exceeds the range of float64.
		return strconv.ParseFloat(v.String(), 64)
	case float64:
		return v, nil
	case []any:
		vec := vals.EmptyList
		for _, elem := range v {
			converted, err := fromJSONInterface(elem)
			if err != nil {
				return nil, err
			}
			vec = vec.Conj(converted)
		}
		return vec, nil
	case map[string]any:
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

func toLines(fm *Frame, inputs Inputs) error {
	out := fm.ByteOutput()
	var errOut error

	inputs(func(v any) {
		if errOut != nil {
			return
		}
		// TODO: Don't ignore the error.
		_, errOut = fmt.Fprintln(out, vals.ToString(v))
	})
	return errOut
}

func toTerminated(fm *Frame, terminator string, inputs Inputs) error {
	if err := checkTerminator(terminator); err != nil {
		return err
	}

	out := fm.ByteOutput()
	var errOut error
	inputs(func(v any) {
		if errOut != nil {
			return
		}
		_, errOut = fmt.Fprint(out, vals.ToString(v), terminator)
	})
	return errOut
}

func toJSON(fm *Frame, inputs Inputs) error {
	encoder := json.NewEncoder(fm.ByteOutput())

	var errEncode error
	inputs(func(v any) {
		if errEncode != nil {
			return
		}
		errEncode = encoder.Encode(v)
	})
	return errEncode
}

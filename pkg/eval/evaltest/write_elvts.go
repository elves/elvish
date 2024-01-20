package evaltest

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vals"
)

const generatedElvTsBanner = "// Generated, can be overwritten"

var elvtsFileStatus = map[string]int{}

// Writes a .elvts file from invocation of Test.
//
// Each test case is its own transcript session. The function name is used as
// h1, and the comment above a test case in the Go source code is used as h2,
// with a fallback title if there are no comments.
type elvtsWriter struct {
	w            io.Writer
	titleForCode map[string]string
}

func newElvtsWriter() elvtsWriter {
	// Actual test site -> Test/TestWithSetup/TestWithEvalerSetup ->
	// testWithSetup -> this function
	pc, filename, line, _ := runtime.Caller(3)
	elvtsName := filename[:len(filename)-len(filepath.Ext(filename))] + ".elvts"
	var w io.Writer
	w, elvtsFileStatus[elvtsName] = getElvts(elvtsName, elvtsFileStatus[elvtsName])

	funcname := runtime.FuncForPC(pc).Name()
	if i := strings.LastIndexByte(funcname, '.'); i != -1 {
		funcname = funcname[i+1:]
	}
	fmt.Fprintf(w, "# %s #\n", funcname)
	return elvtsWriter{w, parseTestTitles(filename, line)}
}

func (ew elvtsWriter) writeCase(c Case, r result) {
	// Add empty line before title
	fmt.Fprintln(ew.w)

	title := "?"
	// The parseTestTitles heuristics assumes a single piece of code (which is
	// the vast majority anyway).
	if len(c.codes) == 1 && ew.titleForCode[c.codes[0]] != "" {
		title = ew.titleForCode[c.codes[0]]
	}
	fmt.Fprintf(ew.w, "## %s ##\n", title)

	fmt.Fprintf(ew.w, "~> %s\n", c.codes[0])
	for _, line := range c.codes[1:] {
		fmt.Fprintf(ew.w, "   %s\n", line)
	}

	fmt.Fprint(ew.w, stringifyResult(r))
}

func (ew elvtsWriter) Close() error {
	if closer, ok := ew.w.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// 0 for unknown
// 1 for already generated
// 2 for don't overwrite
func getElvts(name string, status int) (io.Writer, int) {
	switch status {
	case 0:
		if mayOverwriteElvts(name) {
			file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				panic(err)
			}
			fmt.Fprintln(file, generatedElvTsBanner)
			fmt.Fprintln(file, "//only-on ignore")
			return file, 1
		} else {
			return io.Discard, 2
		}
	case 1:
		file, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(file)
		return file, 1
	case 2:
		return io.Discard, 2
	}
	panic("bad status")
}

func mayOverwriteElvts(name string) bool {
	file, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	return scanner.Text() == generatedElvTsBanner
}

// Parses a block of code that looks like:
//
//	Test(t,
//		// A title
//		That("code 1")...,
//		That("code 2")...,
//	)
//
// and extracts titles for the test code:
//
//	map[string]string{
//		"code 1": "A title",
//	}
func parseTestTitles(filename string, fromLine int) map[string]string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	m := make(map[string]string)
	reader := bufio.NewScanner(file)
	lineno := 0
	for reader.Scan() {
		lineno++
		if lineno < fromLine {
			continue
		}
		line := strings.TrimSpace(reader.Text())
		if line == ")" {
			break
		}
		if title, ok := parseComment(line); ok {
			// Titled clause; collect all comment lines.
			for reader.Scan() {
				line := strings.TrimSpace(reader.Text())
				if moreTitle, ok := parseComment(line); ok {
					title = title + " " + moreTitle
				} else if code, ok := parseThatCode(line); ok {
					if oldTitle, exists := m[code]; exists {
						fmt.Printf("%s:%d: title for %q: %q (already has %q)\n",
							filepath.Base(filename), lineno, code, title, oldTitle)
					} else {
						m[code] = title
					}
					break
				} else {
					break
				}
			}
		}
	}
	if reader.Err() != nil {
		panic(reader.Err())
	}
	return m
}

// "// comment" -> "comment", true
// Everything else -> ?, false
func parseComment(line string) (string, bool) {
	return strings.CutPrefix(line, "// ")
}

var (
	doubleQuoted = regexp.MustCompile(`^"(?:[^\\"]|\\")*"`)
	backQuoted   = regexp.MustCompile("^`[^`]*`")
)

// `That("code")...` -> "code", true
// "That(`code`)..." -> "code", true
// Everything else -> "", false
func parseThatCode(line string) (string, bool) {
	// TODO
	if rest, ok := strings.CutPrefix(line, "That("); ok {
		if dq := doubleQuoted.FindString(rest); dq != "" {
			return strings.ReplaceAll(dq[1:len(dq)-1], `\"`, `"`), true
		} else if bq := backQuoted.FindString(rest); bq != "" {
			return bq[1 : len(bq)-1], true
		}
	}
	return "", false
}

func stringifyResult(r result) string {
	var sb strings.Builder
	for _, value := range r.ValueOut {
		sb.WriteString(valuePrefix + vals.ReprPlain(value) + "\n")
	}
	sb.Write(stripSGR(r.BytesOut))
	sb.Write(stripSGR(r.StderrOut))

	if r.CompilationError != nil {
		sb.WriteString(stripSGRString(r.CompilationError.(diag.Shower).Show("")) + "\n")
	}
	if r.Exception != nil {
		sb.WriteString(stripSGRString(r.Exception.(diag.Shower).Show("")) + "\n")
	}

	return sb.String()
}

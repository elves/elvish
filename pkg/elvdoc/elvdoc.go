// Package elvdoc extracts doc comments of Elvish variables and functions.
package elvdoc

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"strings"

	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/strutil"
)

// Docs records doc comments.
type Docs struct {
	Fns  []Entry
	Vars []Entry
}

// Entry stores the elvdoc for a particular symbol.
type Entry struct {
	Name    string
	HTMLID  string // ID to use in HTML. If empty, just use Name.
	Content string
	LineNo  int // 1-based line number for the first line of Content. 0 if Content is empty.
	Fn      *Fn // nil for variables or functions declared with doc:fn.
}

// Returns e.Content, prepended with function usage if applicable.
func (e Entry) FullContent() string {
	if e.Fn == nil {
		return e.Content
	}
	return fmt.Sprintf("```elvish\n%s\n```\n\n%s", e.Fn.Usage, e.Content)
}

// Fn stores fn-specific information.
type Fn struct {
	// The signature without surrounding pipes, like "a @b".
	Signature string
	// Usage information converted from the original declaration. The function
	// name is qualified, like "ns:f $a $b...".
	Usage string
}

// ExtractFS extracts elvdocs of all modules found under fsys, and returns a map
// from the symbol prefix of a module ("" for builtin, "$mod:" for any other $mod).
//
// See [ExtractFromFS] for how modules correspond to files.
func ExtractAllFromFS(fsys fs.FS) (map[string]Docs, error) {
	prefixes := []string{"", "edit:"}
	modDirs, _ := fs.ReadDir(fsys, "mods")
	for _, modDir := range modDirs {
		if modDir.IsDir() {
			prefixes = append(prefixes, modDir.Name()+":")
		}
	}

	prefixToDocs := map[string]Docs{}
	for _, prefix := range prefixes {
		var err error
		prefixToDocs[prefix], err = ExtractFromFS(fsys, prefix)
		if err != nil {
			return nil, err
		}
	}
	return prefixToDocs, nil
}

// ExtractFromFS extracts elvdoc of a module from fsys. The symbolPrefix is used
// to look up which files to read:
//
//   - "": eval/*.elv (the builtin module)
//   - "edit:": edit/*.elv
//   - "$mod:": mods/$symbolPrefix/*.elv
//
// If symbolPrefix is not empty and doesn't end in ":", this function panics.
func ExtractFromFS(fsys fs.FS, symbolPrefix string) (Docs, error) {
	var subdir string
	switch symbolPrefix {
	case "":
		subdir = "eval"
	case "edit:":
		subdir = "edit"
	default:
		if !strings.HasSuffix(symbolPrefix, ":") {
			panic("symbolPrefix must be empty or ends in :")
		}
		subdir = "mods/" + symbolPrefix[:len(symbolPrefix)-1]
	}
	filenames, err := fs.Glob(fsys, subdir+"/*.elv")
	if err != nil {
		return Docs{}, err
	}
	// Prepare to concatenate subDir/*.elv into one [io.Reader] to pass to
	// [Extract].
	var readers []io.Reader
	for _, filename := range filenames {
		file, err := fsys.Open(filename)
		if err != nil {
			return Docs{}, err
		}
		// Insert an empty line between adjacent files so that the comment
		// block at the end of one file doesn't get merged with the comment
		// block at the start of the next file.
		readers = append(readers, file, strings.NewReader("\n\n"))
	}
	return Extract(io.MultiReader(readers...), symbolPrefix)
}

const (
	singleQuoted = `'(?:[^']|'')*'`
	doubleQuoted = `"(?:[^\\"]|\\.)*"`
	// Bareword, single-quoted and double-quoted. The bareword pattern covers
	// more than what's allowed in Elvish syntax, but that's OK as the context
	// we'll use it in requires a string literal.
	stringLiteralGroup = `([^ '"]+|` + singleQuoted + `|` + doubleQuoted + `)`
	// Any run of non-pipe non-quote runes, or quoted strings. Again this covers
	// a superset of what's allowed, but that's OK.
	signatureGroup = `((?:[^|'"]|` + singleQuoted + `|` + doubleQuoted + `)*)`
)

var (
	// Groups:
	// 1. Name
	// 2. Signature (part inside ||)
	//
	// TODO: Support multi-line function signatures.
	fnRegexp = regexp.MustCompile(`^fn +` + stringLiteralGroup + ` +\{(?: *\|` + signatureGroup + `\|)?`)
	// Groups:
	// 1. Name
	varRegexp = regexp.MustCompile(`^var +` + stringLiteralGroup)
	// Groups:
	// 1. Name
	idRegexp = regexp.MustCompile(`^#doc:id +(.+)`)
)

const showUnstable = "#doc:show-unstable"

// Keeps the state of the current elvdoc block.
//
// An elvdoc block contains a number of consecutive comment lines, followed
// optionally by directive lines (#doc:id or #doc:show-unstable), and ends
// with a fn/var/#doc:fn line.
type blockState struct {
	lines         []string
	startLineNo   int
	seenDirective bool
	id            string
	showUnstable  bool
}

// Uses the state to set relevant fields in the Entry, and resets the state.
func (b *blockState) finish(e *Entry) {
	e.HTMLID = b.id
	e.Content = strutil.JoinLines(b.lines)
	e.LineNo = b.startLineNo
	*b = blockState{}
}

// Extract extracts the elvdoc of one module from an Elvish source.
func Extract(r io.Reader, symbolPrefix string) (Docs, error) {
	var docs Docs
	var block blockState
	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		if line == "#" || strings.HasPrefix(line, "# ") {
			if block.seenDirective {
				return Docs{}, fmt.Errorf("line %d: comment may not follow #doc: directive", lineNo)
			}
			if len(block.lines) == 0 {
				block.startLineNo = lineNo
			}
			if line == "#" {
				block.lines = append(block.lines, "")
			} else {
				block.lines = append(block.lines, line[2:])
			}
		} else if line == showUnstable {
			block.showUnstable = true
			block.seenDirective = true
		} else if m := idRegexp.FindStringSubmatch(line); m != nil {
			block.id = m[1]
			block.seenDirective = true
		} else if m := fnRegexp.FindStringSubmatch(line); m != nil {
			name, sig := unquote(m[1]), m[2]
			qname := symbolPrefix + name
			usage := fnUsage(qname, sig)
			if block.showUnstable || !unstable(name) {
				entry := Entry{Name: qname, Fn: &Fn{sig, usage}}
				block.finish(&entry)
				docs.Fns = append(docs.Fns, entry)
			}
		} else if m := varRegexp.FindStringSubmatch(line); m != nil {
			name := unquote(m[1])
			if block.showUnstable || !unstable(name) {
				entry := Entry{Name: "$" + symbolPrefix + name}
				block.finish(&entry)
				docs.Vars = append(docs.Vars, entry)
			}
		} else {
			block = blockState{}
		}
	}

	return docs, scanner.Err()
}

func unquote(s string) string {
	pn := &parse.Primary{}
	// TODO: Handle error
	parse.ParseAs(parse.Source{Code: s}, pn, parse.Config{})
	return pn.Value
}

func fnUsage(name, sig string) string {
	var sb strings.Builder
	sb.WriteString(parse.QuoteCommandName(name))
	for _, field := range sigFields(sig) {
		sb.WriteByte(' ')
		if strings.HasPrefix(field, "&") {
			sb.WriteString(field)
		} else if strings.HasPrefix(field, "@") {
			sb.WriteString("$" + field[1:] + "...")
		} else {
			sb.WriteString("$" + field)
		}
	}
	return sb.String()
}

func sigFields(sig string) []string {
	pn := &parse.Primary{}
	// TODO: Handle error
	parse.ParseAs(parse.Source{Code: "{|" + sig + "|}"}, pn, parse.Config{})
	var fields []string
	for _, n := range parse.Children(pn) {
		if _, isSep := n.(*parse.Sep); isSep {
			continue
		}
		s := strings.TrimSpace(parse.SourceText(n))
		if s != "" {
			fields = append(fields, s)
		}
	}
	return fields
}

func unstable(s string) bool { return s != "-" && strings.HasPrefix(s, "-") }

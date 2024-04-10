// Package elvdoc extracts doc comments of Elvish variables and functions.
//
// An elvdoc is a continuous sequence of comment lines that consist of:
//
//  1. An optional sequence of "directive" lines, which start with "#" followed
//     immediately by a non-space character.
//
//  2. Any number of content lines, which are either a lone "#", or starts with
//     "# ".
//
// Elvdocs can appear before "var" or "fn" declarations, or at the top of the
// module.
//
// There is one directive recognized by this package, "#doc:show-unstable".
// Normally, symbols starting with "-" are ignored by this package, but adding
// this directive suppresses that behavior. All other directives are left
// unprocessed and returned.
//
// This package doesn't require the content lines to follow any syntax, but
// other packages like pkg/mods/doc and the website generator assume them to be
// Markdown.
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
	File *FileEntry
	Fns  []Entry
	Vars []Entry
}

// FileEntry stores the file-level elvdoc.
type FileEntry struct {
	Directives []string
	Content    string
	LineNo     int
}

// Entry stores the elvdoc for a particular symbol.
type Entry struct {
	Name       string
	Directives []string
	Content    string
	LineNo     int // 1-based line number for the first line of Content. 0 if Content is empty.
	Fn         *Fn // Function-specific information, nil for variables.
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

// ExtractAllFromFS extracts elvdocs of all modules found under fsys, and returns a map
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
			return nil, fmt.Errorf("extracting for prefix %q: %w", prefix, err)
		}
	}
	return prefixToDocs, nil
}

// ExtractFromFS extracts elvdoc of a module from fsys. The symbolPrefix is used
// to look up which files to read:
//
//   - "": eval/*.elv (the builtin module)
//   - "edit:": edit/*.elv
//   - "$mod:": mods/$mod/*.elv
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
)

const showUnstable = "#doc:show-unstable"

// Keeps the state of the current elvdoc block.
//
// An elvdoc block contains a number of consecutive comment lines, followed
// optionally by directive lines (#doc:html-id or #doc:show-unstable), and ends
// with a fn/var/#doc:fn line.
type blockState struct {
	directives   []string
	content      []string
	startLineNo  int
	showUnstable bool
}

// Uses the state to set relevant fields in the Entry, and resets the state.
func (b *blockState) finish() (directives []string, content string, lineNo int, showUnstable bool) {
	directives, content, lineNo, showUnstable = b.directives, strutil.JoinLines(b.content), b.startLineNo, b.showUnstable
	*b = blockState{}
	return
}

// Extract extracts the elvdoc of one module from an Elvish source.
func Extract(r io.Reader, symbolPrefix string) (Docs, error) {
	var docs Docs
	var block blockState
	scanner := bufio.NewScanner(r)
	lineNo := 0
	maybeSetFileEntry := func() {
		// This is a somewhat simplistic criteria for "top of the file", but
		// it's good enough for now.
		if len(docs.Vars) == 0 && len(docs.Fns) == 0 {
			if len(block.directives) > 0 || len(block.content) > 0 {
				directives, content, lineNo, _ := block.finish()
				docs.File = &FileEntry{directives, content, lineNo}
			}
		}
	}
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		if line == "#" || strings.HasPrefix(line, "# ") {
			if len(block.content) == 0 {
				block.startLineNo = lineNo
			}
			if line == "#" {
				block.content = append(block.content, "")
			} else {
				block.content = append(block.content, line[2:])
			}
		} else if strings.HasPrefix(line, "#") {
			if len(block.content) > 0 {
				return Docs{}, fmt.Errorf("line %d: directive must appear at top of elvdoc block", lineNo)
			}
			if line == showUnstable {
				block.showUnstable = true
			} else {
				block.directives = append(block.directives, line[1:])
			}
		} else if m := fnRegexp.FindStringSubmatch(line); m != nil {
			name, sig := unquote(m[1]), m[2]
			qname := symbolPrefix + name
			usage := fnUsage(qname, sig)
			directives, content, lineNo, showUnstable := block.finish()
			if showUnstable || !unstable(name) {
				docs.Fns = append(docs.Fns,
					Entry{qname, directives, content, lineNo, &Fn{sig, usage}})
			}
		} else if m := varRegexp.FindStringSubmatch(line); m != nil {
			name := unquote(m[1])
			directives, content, lineNo, showUnstable := block.finish()
			if showUnstable || !unstable(name) {
				docs.Vars = append(docs.Vars,
					Entry{"$" + symbolPrefix + name, directives, content, lineNo, nil})
			}
		} else {
			maybeSetFileEntry()
			block = blockState{}
		}
	}
	maybeSetFileEntry()

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

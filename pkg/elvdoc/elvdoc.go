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

var (
	// Groups:
	// 1. Name
	// 2. Signature (part inside ||)
	//
	// TODO: Handle more complex cases:
	//
	//  - Quoted |
	//  - Multi-line signature
	fnRegexp = regexp.MustCompile(`^fn +([^ ]+) +\{(?:\|([^|]*)\|)?`)
	// Groups:
	// 1. Name
	varRegexp = regexp.MustCompile(`^var +([^ ]+)`)
	// Groups:
	// 1. Names
	fnNoSigRegexp = regexp.MustCompile(`^#doc:fn +(.+)`)
	// Groups:
	// 1. Name
	idRegexp = regexp.MustCompile(`^#doc:id +(.+)`)
)

const showUnstable = "#doc:show-unstable"

// Extract extracts the elvdoc of one module from an Elvish source.
func Extract(r io.Reader, symbolPrefix string) (Docs, error) {
	var docs Docs
	var block docBlock
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "#" {
			block.lines = append(block.lines, "")
		} else if strings.HasPrefix(line, "# ") {
			block.lines = append(block.lines, line[2:])
		} else if line == showUnstable {
			block.showUnstable = true
		} else if m := fnRegexp.FindStringSubmatch(line); m != nil {
			name, sig := m[1], m[2]
			qname := symbolPrefix + name
			usage := fnUsage(qname, sig)
			id, content, showUnstable := block.consume()
			if showUnstable || !unstable(name) {
				docs.Fns = append(docs.Fns, Entry{qname, id, content, &Fn{sig, usage}})
			}
		} else if m := varRegexp.FindStringSubmatch(line); m != nil {
			name := m[1]
			id, content, showUnstable := block.consume()
			if showUnstable || !unstable(name) {
				docs.Vars = append(docs.Vars, Entry{"$" + symbolPrefix + name, id, content, nil})
			}
		} else if m := fnNoSigRegexp.FindStringSubmatch(line); m != nil {
			name := m[1]
			id, content, _ := block.consume()
			docs.Fns = append(docs.Fns, Entry{symbolPrefix + name, id, content, nil})
		} else if m := idRegexp.FindStringSubmatch(line); m != nil {
			block.id = m[1]
		} else {
			block.consume()
		}
	}

	return docs, scanner.Err()
}

func fnUsage(name, sig string) string {
	var sb strings.Builder
	sb.WriteString(name)
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

type docBlock struct {
	id           string
	lines        []string
	showUnstable bool
}

func (db *docBlock) consume() (id, content string, showUnstable bool) {
	id, content, showUnstable = db.id, strutil.JoinLines(db.lines), db.showUnstable
	*db = docBlock{"", db.lines[:0], false}
	return
}

func unstable(s string) bool { return s != "-" && strings.HasPrefix(s, "-") }

// Package elvdoc extracts doc comments of Elvish variables and functions.
package elvdoc

import (
	"bufio"
	"io"
	"io/fs"
	"regexp"
	"strings"

	"src.elv.sh/pkg/parse"
)

// Docs records doc comments.
type Docs struct {
	Fns  []Entry
	Vars []Entry
}

// Entry is a doc comment entry.
type Entry struct {
	Name string
	// ID to use in HTML
	ID      string
	Content string
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
			content := fnUsage(symbolPrefix+name, sig)
			id := ""
			showUnstable := false
			if len(block.lines) > 0 {
				var blockContent string
				id, blockContent, showUnstable = block.consume()
				content += "\n" + blockContent
			}
			if showUnstable || !unstable(name) {
				docs.Fns = append(docs.Fns, Entry{name, id, content})
			}
		} else if m := varRegexp.FindStringSubmatch(line); m != nil {
			name := m[1]
			id, content, showUnstable := block.consume()
			if showUnstable || !unstable(name) {
				docs.Vars = append(docs.Vars, Entry{name, id, content})
			}
		} else if m := fnNoSigRegexp.FindStringSubmatch(line); m != nil {
			name := m[1]
			id, content, _ := block.consume()
			docs.Fns = append(docs.Fns, Entry{name, id, content})
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
	sb.WriteString("```elvish\n")
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
	sb.WriteString("\n```\n")
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
	id = db.id
	db.id = ""

	var sb strings.Builder
	for _, line := range db.lines {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	db.lines = db.lines[:0]

	showUnstable = db.showUnstable
	db.showUnstable = false

	return id, sb.String(), showUnstable
}

func unstable(s string) bool { return s != "-" && strings.HasPrefix(s, "-") }

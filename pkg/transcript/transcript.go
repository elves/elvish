// Package transcript contains utilities for working with Elvish transcripts.
//
// # Basic syntax
//
// In its most basic form, a transcript consists of a series of code entered
// after a prompt, each followed by the resulting output:
//
//	~> echo foo
//	foo
//	~> echo lorem
//	   echo ipsum
//	lorem
//	ipsum
//
// A line starting with a prompt (as defined by [PromptPattern]) is considered
// to start code; code extends to further lines that are indented to align with
// the prompt. The other lines are considered output.
//
// # Headings and sessions
//
// Two levels of headings are supported: "# h1 #" and "## h2 ##". They split a
// transcript into a tree of multiple sessions, and the titles become their
// names.
//
// For example, suppose that a.elvts contains the following content:
//
//	~> echo hello
//	hello
//
//	# foo #
//
//	~> foo
//	something is done
//
//	# bar #
//
//	## 1 ##
//	~> bar 1
//	something is 1 done
//
//	## 2 ##
//	~> bar 2
//	something is 2 done
//
// This file contains the following tree:
//
//	a.elvts
//		foo
//		bar
//			1
//			2
//
// Leading and trailing empty lines are stripped from a session, but internal
// empty lines are kept intact. This also applies to transcripts with no
// headings (and thus consisting of exactly one session).
//
// # Comments and directives
//
// A line starting with "// " or consisting of 2 or more "/"s and nothing else
// is a comment. Comments are ignored and can appear anywhere, except that they
// can't interrupt multi-line code.
//
// A line starting with "//" but is not a comment is a directive. Directives can
// only appear at the beginning of a session, possibly after other directives,
// comments or empty lines.
//
// # Sessions in .elv files
//
// An .elv file may contain elvdocs for their variables or functions, which in
// turn may contain examples given as elvish-transcript code blocks.
//
// Each of those code block is considered a transcript, named
// $filename/$symbol/$name, where $name is the additional words after the
// "elvish-transcript" in the opening fence, defaulting to an empty string.
//
// File-level directives and symbol-level directives starting with "#//" are
// supported.
//
// As an example, suppose a.elv contains the following content:
//
//	#//dir1
//
//	#//dir2
//	# Does something.
//	#
//	# Example:
//	#
//	# ```elvish-transcript
//	# ~> foo
//	# something is done
//	# ```
//	fn foo {|| }
//
//	# Does something depending on argument.
//	#
//	# Example:
//	#
//	# ```elvish-transcript 1
//	# ~> bar 1
//	# something 1 is done
//	# ```
//	#
//	# Another example:
//	#
//	# ```elvish-transcript 2
//	# ~> bar 2
//	# something 2 is done
//	# ```
//	fn bar {|x| }
//
// This creates the following tree:
//
//	a.elv
//		foo
//			unnamed
//		bar
//			1
//			2
//
// These transcripts can also contain headings, which split them into further
// smaller sessions.
package transcript

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/strutil"
)

// Node is the result of parsing transcripts. It can represent an .elvts file, a
// elvish-transcript block within the elvdoc of an .elv file, or an section
// within them started by a header.
type Node struct {
	Name         string
	Directives   []string
	Interactions []Interaction
	Children     []*Node
	// [LineFrom, LineTo)
	LineFrom, LineTo int
}

// ParseFromFS scans fsys recursively for .elv and .elvts files, and
// extract transcript sessions from them.
func ParseFromFS(fsys fs.FS) ([]*Node, error) {
	var nodes []*Node
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		parseNode := Parse
		switch filepath.Ext(path) {
		case ".elv":
			parseNode = parseElv
		case ".elvts":
		default:
			return nil
		}
		file, err := fsys.Open(path)
		if err != nil {
			return err
		}
		node, err := parseNode(path, file)
		if err != nil {
			return err
		}
		nodes = append(nodes, node)
		return nil
	})
	return nodes, err
}

func readAllLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Scans the elvdoc in an Elvish source file for elvish-transcript blocks and
// parses each one similar to an .elvts file. Each block becomes a [Node] on its
// own, named like "foo.elv/symbol/code-fence-info" or "foo.elv/symbol" (if
// fence info is empty).
func parseElv(filename string, r io.Reader) (*Node, error) {
	docs, err := elvdoc.Extract(r, "")
	if err != nil {
		return nil, fmt.Errorf("parse %s for elvdoc: %w", filename, err)
	}
	fileNode := &Node{
		Name:     filename,
		LineFrom: 1,
		// LineTo will be updated as children get added
	}
	if docs.File != nil {
		fileNode.Directives = testDirectivesFromElvdoc(docs.File.Directives)
	}

	parseEntries := func(entries []elvdoc.Entry) error {
		for _, entry := range entries {
			codec := transcriptExtractor{filename, entry.LineNo - 1, nil, nil}
			md.Render(entry.Content, &codec)
			if codec.err != nil {
				return codec.err
			}
			symbolNode := &Node{
				Name:       entry.Name,
				Directives: testDirectivesFromElvdoc(entry.Directives),
				Children:   codec.nodes,
				LineFrom:   entry.LineNo,
				LineTo:     entry.LineNo + strings.Count(entry.Content, "\n"),
			}
			fileNode.Children = append(fileNode.Children, symbolNode)
			if fileNode.LineTo < symbolNode.LineTo {
				fileNode.LineTo = symbolNode.LineTo
			}
		}
		return nil
	}
	err = parseEntries(docs.Fns)
	if err != nil {
		return nil, err
	}
	err = parseEntries(docs.Vars)
	if err != nil {
		return nil, err
	}
	return fileNode, nil
}

func testDirectivesFromElvdoc(directives []string) []string {
	var testDirectives []string
	for _, directive := range directives {
		if testDirective, ok := strings.CutPrefix(directive, "//"); ok {
			testDirectives = append(testDirectives, testDirective)
		}
	}
	return testDirectives
}

// A [md.Codec] implementation that extracts elvish-transcript code blocks from
// an elvdoc block as sessions.
type transcriptExtractor struct {
	filename     string
	lineNoOffset int

	nodes []*Node
	err   error
}

func (e *transcriptExtractor) Do(op md.Op) {
	if e.err != nil {
		return
	}
	if op.Type == md.OpCodeBlock {
		if lang, name, _ := strings.Cut(op.Info, " "); lang == "elvish-transcript" {
			// The first line of the code block is the fence line, add 1 to get
			// the first line of the actual content.
			lineNo := e.lineNoOffset + op.LineNo + 1
			node, err := parseNode(name, fileLines{e.filename, op.Lines, lineNo})
			if err != nil {
				e.err = err
				return
			}
			e.nodes = append(e.nodes, node)
		}
	}
}

// Parse parses the transcript sessions from an .elvts file.
func Parse(path string, r io.Reader) (*Node, error) {
	lines, err := readAllLines(r)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return parseNode(path, fileLines{path, lines, 1})
}

// Represents a range of lines from a file.
type fileLines struct {
	filename    string
	lines       []string
	startLineNo int // line number of lines[0]
}

func (fl *fileLines) describeLine(i int) string {
	return fmt.Sprintf("%s:%d", fl.filename, fl.lineNo(i))
}

func (fl *fileLines) lineNo(i int) int { return i + fl.startLineNo }

func (fl *fileLines) slice(i, j int) fileLines {
	return fileLines{fl.filename, fl.lines[i:j], fl.lineNo(i)}
}

// Parses a single node. This could be an .elvts file, or part of an .elv file.
func parseNode(name string, fl fileLines) (*Node, error) {
	// Path from root to current node. Index corresponds to level, so
	// nodeStack[0] is the root,  nodeStack[1] is the currently active h1, and
	// so on.
	nodeStack := []*Node{{Name: name, LineFrom: fl.lineNo(0)}}

	for i := 0; i < len(fl.lines); {
		if title, level, ok := parseHeading(fl.lines[i]); ok {
			// Consume a heading line. This branch will always be entered with
			// the possible exception of the first iteration, because the
			// condition is the terminating condition of the loop below to find
			// which lines to parse as a Session.
			if level > len(nodeStack) {
				return nil, fmt.Errorf("%s: h%d before h%d", fl.describeLine(i), level, level-1)
			}
			node := &Node{Name: title, LineFrom: fl.lineNo(i)}
			parent := nodeStack[level-1]
			parent.Children = append(parent.Children, node)
			// Terminate all nodes that are at the new node's level or a deeper
			// level.
			for _, n := range nodeStack[level:] {
				n.LineTo = fl.lineNo(i)
			}
			// Remove terminated nodes and push new node.
			nodeStack = append(nodeStack[:level], node)
			// We're done with this line.
			i++
		}
		// Consume all lines to the next heading, and parse it as a Session to
		// attach to the current node.
		var j int
		for j = i + 1; j < len(fl.lines); j++ {
			if _, _, isHeading := parseHeading(fl.lines[j]); isHeading {
				break
			}
		}
		err := parseSession(nodeStack[len(nodeStack)-1], fl.slice(i, j))
		i = j
		if err != nil {
			return nil, err
		}
	}
	// Nodes that are still active now terminate at the EOF.
	for _, n := range nodeStack {
		n.LineTo = fl.lineNo(len(fl.lines))
	}
	return nodeStack[0], nil
}

func parseHeading(line string) (title string, level int, ok bool) {
	if strings.HasPrefix(line, "# ") && strings.HasSuffix(line, " #") {
		return line[2 : len(line)-2], 1, true
	} else if strings.HasPrefix(line, "## ") && strings.HasSuffix(line, " ##") {
		return line[3 : len(line)-3], 2, true
	} else if strings.HasPrefix(line, "### ") && strings.HasSuffix(line, " ###") {
		return line[4 : len(line)-4], 3, true
	} else {
		return "", 0, false
	}
}

// Interaction represents a single REPL interaction - user input followed by the
// shell's output. Prompt is never empty.
type Interaction struct {
	Prompt string
	Code   string
	// [CodeLineFrom, CodeLineTo) identifies the range of code lines.
	CodeLineFrom, CodeLineTo int

	Output string
	// [OutputLineFrom, OutputlineTo) identifies the range of output lines,
	// excluding any leading and trailing comment lines.
	OutputLineFrom, OutputLineTo int
}

// PromptAndCode returns prompt and code concatenated, with spaces prepended to
// continuation lines in Code to align with the first line.
func (i Interaction) PromptAndCode() string {
	if i.Code == "" {
		return i.Prompt
	}
	lines := strings.Split(i.Code, "\n")
	var sb strings.Builder
	sb.WriteString(i.Prompt + lines[0])
	continuation := strings.Repeat(" ", len(i.Prompt))
	for _, line := range lines[1:] {
		sb.WriteString("\n" + continuation + line)
	}
	return sb.String()
}

// PromptPattern defines how to match prompts, used to determine which lines
// start the code part of an interaction.
var PromptPattern = regexp.MustCompile(`^[~/][^ ]*> `)

var (
	errFirstLineDoesntHavePrompt            = errors.New("first non-comment line of a session doesn't have prompt")
	errDirectiveOnlyAllowedAtStartOfSession = errors.New("directive only allowed at start of a session")
)

// Parses a session into n. Mutates n.Directives and n.Interactions on success.
func parseSession(n *Node, fl fileLines) error {
	lines := fl.lines
	// Process leading empty lines, comment lines and directive lines.
	var directives []string
	start := 0
	for ; start < len(lines); start++ {
		if lines[start] == "" || isComment(lines[start]) {
			// do nothing
		} else if directive, ok := parseDirective(lines[start]); ok {
			directives = append(directives, directive)
		} else {
			break
		}
	}
	if start < len(lines) && !PromptPattern.MatchString(lines[start]) {
		return fmt.Errorf("%s: %w", fl.describeLine(start), errFirstLineDoesntHavePrompt)
	}
	// Remove trailing empty lines and comment lines.
	for len(lines) > 0 && (lines[len(lines)-1] == "" || isComment(lines[len(lines)-1])) {
		lines = lines[:len(lines)-1]
	}
	// Parse interactions.
	var interactions []Interaction
	for i := start; i < len(lines); {
		codeLineFrom := fl.lineNo(i)
		// Consume the first code line.
		prompt := PromptPattern.FindString(lines[i])
		code := []string{lines[i][len(prompt):]}
		i++
		// Consume continuation code lines.
		continuation := strings.Repeat(" ", len(prompt))
		for i < len(lines) && strings.HasPrefix(lines[i], continuation) {
			code = append(code, lines[i][len(continuation):])
			i++
		}
		codeLineTo := fl.lineNo(i)

		// Ignore comment lines between code and output.
		for i < len(lines) && isComment(lines[i]) {
			i++
		}
		// Consume output lines, ignoring internal and trailing comment lines.
		var output []string
		outputLineFrom := fl.lineNo(i)
		outputLineTo := fl.lineNo(i)
		for i < len(lines) && !PromptPattern.MatchString(lines[i]) {
			if _, ok := parseDirective(lines[i]); ok {
				return fmt.Errorf("%s: %w",
					fl.describeLine(i), errDirectiveOnlyAllowedAtStartOfSession)
			} else if isComment(lines[i]) {
				// Do nothing
			} else {
				output = append(output, lines[i])
				outputLineTo = fl.lineNo(i + 1)
			}
			i++
		}
		interactions = append(interactions, Interaction{
			prompt,
			// Code doesn't include the trailing newline, so a simple
			// strings.Join is appropriate.
			strings.Join(code, "\n"), codeLineFrom, codeLineTo,
			strutil.JoinLines(output), outputLineFrom, outputLineTo})
	}
	n.Directives = directives
	n.Interactions = interactions
	return nil
}

var slashOnlyCommentPattern = regexp.MustCompile(`^///*$`)

func isComment(line string) bool {
	return strings.HasPrefix(line, "// ") || slashOnlyCommentPattern.MatchString(line)
}

func parseDirective(line string) (string, bool) {
	if strings.HasPrefix(line, "//") && !isComment(line) {
		return line[2:], true
	}
	return "", false
}

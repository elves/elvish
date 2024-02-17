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
// transcript into multiple sessions and are used to name them.
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
// This file contains three sessions: a.elvts, a.elvts/foo, a.elvts/bar/1 and
// a.elvts/bar/2.
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
// Directives propagate to "lower-level" sessions. This mechanism is best shown
// with an example:
//
//	//top
//
//	# h1 #
//	//h1
//
//	## h2 ##
//	//h2
//
//	~> echo foo
//	foo
//
// In the "echo foo" session, all of "top", "h1" and "h2" directives are active.
//
// # Sessions in .elv files
//
// An .elv file may contain elvdocs for their variables or functions, which in
// turn may contain examples given as elvish-transcript code blocks.
//
// Each of those code block is considered a transcript, named $filename/$symbol.
// If there are additional words after the "elvish-transcript" in the opening
// fence, they are appended to name of the transcript, becoming
// $filename/$symbol/$name.
//
// As an example, suppose a.elv contains the following content:
//
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
// This creates three sessions: a.elv/foo, a.elv/bar/1 and a.elv/bar/2.
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
)

// Session represents a REPL session.
type Session struct {
	Name         string
	Directives   []string
	Interactions []Interaction
}

// ParseFromFS scans fsys recursively for .elv and .elvts files, and
// extract transcript sessions from them.
func ParseFromFS(fsys fs.FS) ([]Session, error) {
	var sessions []Session
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		var parseSessions func(namePrefix string, file io.Reader) ([]Session, error)
		switch filepath.Ext(path) {
		case ".elv":
			parseSessions = parseElvdocInElv
		case ".elvts":
			parseSessions = parseElvts
		default:
			return nil
		}
		file, err := fsys.Open(path)
		if err != nil {
			return err
		}
		moreSessions, err := parseSessions(path, file)
		if err != nil {
			return err
		}
		sessions = append(sessions, moreSessions...)
		return nil
	})
	return sessions, err
}

// Scans the elvdoc in an Elvish source file for elvish-transcript blocks and
// parses each one similar to an .elvts file.
func parseElvdocInElv(namePrefix string, r io.Reader) ([]Session, error) {
	docs, err := elvdoc.Extract(r, "")
	if err != nil {
		return nil, fmt.Errorf("parsing %s for elvdoc: %w", namePrefix, err)
	}
	var sessions []Session
	parseEntries := func(entries []elvdoc.Entry, prefix string) error {
		for _, entry := range entries {
			codec := transcriptExtractor{namePrefix: prefix + entry.Name}
			md.Render(entry.Content, &codec)
			if codec.err != nil {
				return codec.err
			}
			sessions = append(sessions, codec.sessions...)
		}
		return nil
	}
	err = parseEntries(docs.Fns, namePrefix+"/")
	if err != nil {
		return nil, err
	}
	err = parseEntries(docs.Vars, namePrefix+"/$")
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// A [md.Codec] implementation that extracts elvish-transcript code blocks as
// sessions.
type transcriptExtractor struct {
	namePrefix string // $filename/$symbol
	sessions   []Session
	err        error
}

func (e *transcriptExtractor) Do(op md.Op) {
	if e.err != nil {
		return
	}
	if op.Type == md.OpCodeBlock {
		fields := strings.Fields(op.Info)
		if len(fields) > 0 && fields[0] == "elvish-transcript" {
			name := e.namePrefix
			if len(fields) > 1 {
				name += "/" + strings.Join(fields[1:], " ")
			}
			// Ideally we'd like to pass the line number where this code block
			// starts, so that parseSessionsInBlock can report the accurate line
			// number within the overall file (currently it will report
			// something like a.elv/x:12). To do that, we'll need to change
			// pkg/md to keep track of the line number and pass it to the Codec
			// first.
			sessions, err := parseInner(fileLines{name, op.Lines, 1})
			if err != nil {
				e.err = err
				return
			}
			e.sessions = append(e.sessions, sessions...)
		}
	}
}

// Parses an .elvts file.
func parseElvts(name string, r io.Reader) ([]Session, error) {
	lines, err := readAllLines(r)
	if err != nil {
		return nil, err
	}
	return parseInner(fileLines{name, lines, 1})
}

func readAllLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Represents a range of lines from a file.
type fileLines struct {
	filename    string
	lines       []string
	startLineno int // line number of lines[0]
}

func (fl *fileLines) describeLine(i int) string {
	return fmt.Sprintf("%s:%d", fl.filename, i+fl.startLineno)
}

type section struct {
	heading    string
	directives []string
}

func parseInner(fl fileLines) ([]Session, error) {
	var sessions []Session
	sectionStack := []section{{fl.filename, nil}}
	var testLines []string
	var testLinesStartLineno int
	consumeTestLines := func() error {
		name := joinHeading(sectionStack)
		session, err := parseSession(name, fileLines{fl.filename, testLines, testLinesStartLineno})
		testLines = nil
		if err != nil {
			return err
		}
		sectionStack[len(sectionStack)-1].directives = session.Directives
		if len(session.Interactions) > 0 {
			sessions = append(sessions, Session{
				name, joinDirectives(sectionStack), session.Interactions})
		}
		return nil
	}
	for i, line := range fl.lines {
		if strings.HasPrefix(line, "# ") && strings.HasSuffix(line, " #") {
			err := consumeTestLines()
			if err != nil {
				return nil, err
			}
			sectionStack = append(sectionStack[:1], section{heading: line[2 : len(line)-2]})
		} else if strings.HasPrefix(line, "## ") && strings.HasSuffix(line, " ##") {
			err := consumeTestLines()
			if err != nil {
				return nil, err
			}
			if len(sectionStack) < 2 {
				return nil, fmt.Errorf("%s: h2 before any h1", fl.describeLine(i))
			}
			sectionStack = append(sectionStack[:2], section{heading: line[3 : len(line)-3]})
		} else {
			if len(testLines) == 0 {
				testLinesStartLineno = i + fl.startLineno
			}
			testLines = append(testLines, line)
		}
	}
	err := consumeTestLines()
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func joinHeading(s []section) string {
	headings := make([]string, len(s))
	for i, section := range s {
		headings[i] = section.heading
	}
	return strings.Join(headings, "/")
}

func joinDirectives(s []section) []string {
	var ds []string
	for _, section := range s {
		ds = append(ds, section.directives...)
	}
	return ds
}

// Interaction represents a single REPL interaction - user input followed by the
// shell's output. Prompt is never empty.
type Interaction struct {
	Prompt string
	Code   string
	Output string
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

func parseSession(name string, fl fileLines) (Session, error) {
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
		return Session{}, fmt.Errorf("%s: %w", fl.describeLine(start), errFirstLineDoesntHavePrompt)
	}
	// Remove trailing empty lines and comment lines.
	for len(lines) > 0 && (lines[len(lines)-1] == "" || isComment(lines[len(lines)-1])) {
		lines = lines[:len(lines)-1]
	}
	// Parse interactions.
	var interactions []Interaction
	for i := start; i < len(lines); {
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
		// Consume output lines, ignoring comment lines.
		var output []string
		for i < len(lines) && !PromptPattern.MatchString(lines[i]) {
			if _, ok := parseDirective(lines[i]); ok {
				return Session{}, fmt.Errorf("%s: %w",
					fl.describeLine(i), errDirectiveOnlyAllowedAtStartOfSession)
			} else if isComment(lines[i]) {
			} else {
				output = append(output, lines[i])
			}
			i++
		}
		interactions = append(interactions, Interaction{
			prompt,
			// Code doesn't include the trailing newline, so a simple
			// strings.Join is appropriate.
			strings.Join(code, "\n"),
			joinLines(output)})
	}
	return Session{name, directives, interactions}, nil
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

// Equivalent to appending each line with a "\n" and joining all of them.
func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

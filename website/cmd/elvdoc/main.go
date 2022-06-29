package main

import (
	"bufio"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	run(os.Args[1:], os.Stdin, os.Stdout)
}

func run(args []string, in io.Reader, out io.Writer) {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	var ns = flags.String("ns", "", "namespace prefix")
	var raw = flags.Bool("raw", false, "raw output (for generating help text)")

	err := flags.Parse(args)
	if err != nil {
		log.Fatal(err)
	}
	args = flags.Args()

	if len(args) > 0 {
		extractPaths(args, *ns, *raw, out)
	} else {
		extract(in, *ns, *raw, out)
	}
}

func extractPaths(paths []string, ns string, raw bool, out io.Writer) {
	var files []string
	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			log.Fatal(err)
		}
		if stat.IsDir() {
			goFiles := mustGlob(filepath.Join(path, "*.go"))
			files = append(files, goFiles...)
			elvFiles := mustGlob(filepath.Join(path, "*.elv"))
			files = append(files, elvFiles...)
		} else {
			files = append(files, path)
		}
	}

	reader, cleanup, err := multiFile(files)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
	extract(reader, ns, raw, out)
}

func mustGlob(p string) []string {
	files, err := filepath.Glob(p)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

// Makes a reader that concatenates multiple files.
func multiFile(names []string) (io.Reader, func(), error) {
	readers := make([]io.Reader, len(names))
	closers := make([]io.Closer, len(names))
	for i, name := range names {
		file, err := os.Open(name)
		if err != nil {
			for j := 0; j < i; j++ {
				closers[j].Close()
			}
			return nil, nil, err
		}
		readers[i] = file
		closers[i] = file
	}
	return io.MultiReader(readers...), func() {
		for _, closer := range closers {
			closer.Close()
		}
	}, nil
}

// htmlMarkdownFunc outputs data suitable for embedding Elvish variable and function documentation
// in HTML pages.
func htmlMarkdownFunc(w io.Writer, heading, entryType, prefix string, m map[string]string) {
	fmt.Fprintf(w, "# %s\n", heading)
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Slice(names,
		func(i, j int) bool {
			return symbolForSort(names[i]) < symbolForSort(names[j])
		})
	for _, name := range names {
		fmt.Fprintln(w)
		fullName := prefix + name
		// Create anchors for Docset. These anchors are used to show a ToC;
		// the mkdsidx.py script also looks for those anchors to generate
		// the SQLite index.
		//
		// Some builtin commands are documented together. Create an anchor
		// for each of them.
		for _, s := range strings.Fields(fullName) {
			if strings.HasPrefix(s, "{#") {
				continue
			}
			fmt.Fprintf(w,
				"<a name='//apple_ref/cpp/%s/%s' class='dashAnchor'></a>\n\n",
				entryType, url.QueryEscape(html.UnescapeString(s)))
		}
		if strings.Contains(fullName, "{#") {
			fmt.Fprintf(w, "## %s\n", fullName)
		} else {
			// pandoc strips punctuations from the ID, turning "mod:name"
			// into "modname". Explicitly preserve the original full name
			// by specifying an attribute. We still strip the leading $ for
			// variables since pandoc will treat "{#$foo}" as part of the
			// title.
			id := strings.TrimPrefix(fullName, "$")
			fmt.Fprintf(w, "## %s {#%s}\n", fullName, id)
		}
		// The body is guaranteed to have a trailing newline, hence Fprint
		// instead of Fprintln.
		fmt.Fprint(w, m[name])
	}
}

// rawMarkdownFunc outputs data suitable for embedding Elvish variable and function documentation in
// the Elvish binary for use by the builtin `help` command.
var specialSymsRe = regexp.MustCompile(`(.*) {#[^}]+}$`)

func rawMarkdownFunc(w io.Writer, heading, entryType, prefix string, m map[string]string) {
	_ = heading
	for name, definition := range m {
		fmt.Fprintln(w)
		match := specialSymsRe.FindStringSubmatch(name)
		var fullNames []string
		if len(match) == 0 { // the common case
			fullNames = []string{prefix + name}
		} else { // the uncommon case
			// This is a special-case for a handful of builtin functions, such as `+`, `==`, and
			// `==s`. We ignore the alias provided in the comment (e.g., "add", "num-cmp" and
			// "str-cmp") for generating the output and use the individual special strings.
			for _, name := range strings.Split(match[1], " ") {
				fullNames = append(fullNames, prefix+name)
			}
		}
		if fullNames[0][0] != '$' {
			// TODO: Modify the documentation to use "```elvish-usage" then remove this hack.
			if strings.HasPrefix(definition, "\n```elvish\n") {
				definition = "\n```elvish-usage\n" + definition[len("\n```elvish\n"):]
			}
		}
		// The use of `@@elvdoc` rather than `## elvdoc` is to keep the Markdown to terminal
		// transformer from handling this magic line. In contrast to htmlMarkdownFunc.
		for _, fullName := range fullNames {
			if fullName[0] == '$' {
				fmt.Fprintln(w, "@@elvdoc:var:"+fullName[1:])
			} else {
				fmt.Fprintln(w, "@@elvdoc:fn:"+fullName)
			}
			fmt.Fprint(w, rawMarkupTransforms(definition))
		}
	}
}

// This handles some simple transformations of the Markdown extensions used by the Elvish
// documentation into something friendlier for the `help` command.
func rawMarkupTransforms(definition string) string {
	var result strings.Builder
	result.Grow(len(definition) + 2*len(definition)/10)
	for _, line := range strings.Split(definition, "\n") {
		if strings.HasPrefix(line, "@cf ") {
			line = "See also: " + line[len("@cf "):]
		}
		result.WriteString(line)
		result.WriteByte('\n')
	}
	return result.String()
}

func extract(r io.Reader, ns string, raw bool, w io.Writer) {
	bufr := bufio.NewReader(r)

	fnDocs := make(map[string]string)
	varDocs := make(map[string]string)

	// Reads a block of line comments, i.e. a continuous range of lines that
	// start with a comment leader (// or #). Returns the content, with the
	// comment leader and any spaces after it removed. The content always ends
	// with a newline, even if the last line of the comment is the last line of
	// the file without a trailing newline.
	//
	// Discards the first line after the comment block.
	readCommentBlock := func() (string, error) {
		builder := &strings.Builder{}
		for {
			line, err := bufr.ReadString('\n')
			if err == io.EOF && len(line) > 0 {
				// We pretend that the file always have a trailing newline even
				// if it does not. The next ReadString will return io.EOF again
				// with an empty line.
				line += "\n"
				err = nil
			}
			line, isComment := stripCommentLeader(line)
			if isComment && err == nil {
				// Trim any spaces after the comment leader. The line already
				// has a trailing newline, so no need to write \n.
				builder.WriteString(strings.TrimPrefix(line, " "))
			} else {
				// Discard this line, finalize the builder, and return.
				return builder.String(), err
			}
		}
	}

	for {
		line, err := bufr.ReadString('\n')
		line, isComment := stripCommentLeader(line)

		const (
			varDocPrefix = "elvdoc:var "
			fnDocPrefix  = "elvdoc:fn "
		)

		if isComment && err == nil {
			switch {
			case strings.HasPrefix(line, varDocPrefix):
				varName := line[len(varDocPrefix) : len(line)-1]
				varDocs[varName], err = readCommentBlock()
			case strings.HasPrefix(line, fnDocPrefix):
				fnName := line[len(fnDocPrefix) : len(line)-1]
				fnDocs[fnName], err = readCommentBlock()
			}
		}

		if err != nil {
			if err != io.EOF {
				log.Fatalf("read: %v", err)
			}
			break
		}
	}

	var write func(w io.Writer, heading, entryType, prefix string, m map[string]string)
	if raw {
		write = rawMarkdownFunc
	} else {
		write = htmlMarkdownFunc
	}

	if len(varDocs) > 0 {
		write(w, "Variables", "Variable", "$"+ns, varDocs)
	}
	if len(fnDocs) > 0 {
		if !raw && len(varDocs) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w)
		}
		write(w, "Functions", "Function", ns, fnDocs)
	}
}

func stripCommentLeader(s string) (string, bool) {
	if strings.HasPrefix(s, "#") {
		return s[1:], true
	} else if strings.HasPrefix(s, "//") {
		return s[2:], true
	}
	return s, false
}

var sortSymbol = map[string]string{
	"+": " a",
	"-": " b",
	"*": " c",
	"/": " d",
}

func symbolForSort(s string) string {
	// Hack to sort + - * / in that order.
	if t, ok := sortSymbol[strings.Fields(s)[0]]; ok {
		return t
	}
	// If there is a leading dash, move it to the end.
	if strings.HasPrefix(s, "-") {
		return s[1:] + "-"
	}
	return s
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	run(os.Args[1:], os.Stdin, os.Stdout)
}

func run(args []string, in io.Reader, out io.Writer) {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	var (
		directory = flags.Bool("dir", false, "read from .go files in directories")
		filter    = flags.Bool("filter", false, "act as a Markdown file filter")
		ns        = flags.String("ns", "", "namespace prefix")
	)

	err := flags.Parse(args)
	if err != nil {
		log.Fatal(err)
	}
	args = flags.Args()

	switch {
	case *directory:
		extractDirs(args, *ns, out)
	case *filter:
		// NOTE: Ignores arguments.
		filterMarkdown(in, out)
	case len(args) > 0:
		extractFiles(args, *ns, out)
	default:
		extract(in, *ns, out)
	}
}

const markdownLeader = "$elvdoc "

var emptyReader = &strings.Reader{}

func filterMarkdown(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if arg := strings.TrimPrefix(line, markdownLeader); arg != line {
			args := strings.Fields(arg)
			run(args, emptyReader, out)
		} else {
			fmt.Fprintln(out, line)
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Fatal(err)
	}
}

func extractDirs(dirs []string, ns string, out io.Writer) {
	var files []string
	for _, dir := range dirs {
		files = append(files, goFilesInDirectory(dir)...)
	}
	extractFiles(files, ns, out)
}

func extractFiles(files []string, ns string, out io.Writer) {
	reader, cleanup, err := multiFile(files)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
	extract(reader, ns, out)
}

// Returns all .go files in the given directory.
func goFilesInDirectory(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("walk %v: %v", dir, err)
	}
	var paths []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".go" {
			paths = append(paths, filepath.Join(dir, file.Name()))
		}
	}
	return paths
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

func extract(r io.Reader, ns string, w io.Writer) {
	bufr := bufio.NewReader(r)

	fnDocs := make(map[string]string)
	varDocs := make(map[string]string)

	// Reads a block of line comments, i.e. a continuous range of lines that
	// start with //. Returns the content, with the leading // and any spaces
	// after it removed. The content always ends with a newline, even if the
	// last line of the comment is the last line of the file without a trailing
	// newline.
	//
	// Will discard the first line after the comment block.
	readCommentBlock := func() (string, error) {
		builder := &strings.Builder{}
		for {
			line, err := bufr.ReadString('\n')
			if err == io.EOF && len(line) > 0 {
				// We pretend that the file always have a trailing newline even
				// if it does not exist. The next ReadString will return io.EOF
				// again with an empty line.
				line += "\n"
				err = nil
			}
			if !strings.HasPrefix(line, "//") || err != nil {
				// Discard this line, finalize the builder, and return.
				return builder.String(), err
			}
			// The line already has a trailing newline.
			builder.WriteString(strings.TrimLeft(line[len("//"):], " "))
		}
	}

	for {
		line, err := bufr.ReadString('\n')

		const (
			varDocPrefix = "//elvdoc:var "
			fnDocPrefix  = "//elvdoc:fn "
		)

		if err == nil {
			switch {
			case strings.HasPrefix(line, varDocPrefix):
				varName := line[len(varDocPrefix) : len(line)-1]
				varDocs["$"+ns+varName], err = readCommentBlock()
			case strings.HasPrefix(line, fnDocPrefix):
				fnName := line[len(fnDocPrefix) : len(line)-1]
				fnDocs[ns+fnName], err = readCommentBlock()
			}
		}

		if err != nil {
			if err != io.EOF {
				log.Fatalf("read: %v", err)
			}
			break
		}
	}

	write := func(heading string, m map[string]string) {
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
			fmt.Fprintf(w, "## %s\n", name)
			// The body is guaranteed to have a trailing newline, hence Fprint
			// instead of Fprintln.
			fmt.Fprint(w, m[name])
		}
	}

	if len(varDocs) > 0 {
		write("Variables", varDocs)
	}
	if len(fnDocs) > 0 {
		if len(varDocs) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w)
		}
		write("Functions", fnDocs)
	}
}

func symbolForSort(s string) string {
	// If there is a leading dash, move it to the end.
	if strings.HasPrefix(s, "-") {
		return s[1:] + "-"
	}
	return s
}

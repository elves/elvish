package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	run(os.Args[1:], os.Stdin, os.Stdout)
}

var (
	recursive = flag.Bool("R", false, "recursively read from all .go files")
)

func run(args []string, in io.Reader, out io.Writer) {
	flag.CommandLine.Parse(args)
	args = flag.Args()
	if *recursive {
		var paths []string
		for _, dir := range args {
			morePaths := getAllGoFiles(dir)
			paths = append(paths, morePaths...)
		}
		extractFromFiles(paths, out)
	} else {
		if len(args) == 0 {
			extract(in, out)
		} else {
			extractFromFiles(args, out)
		}
	}
}

func extractFromFiles(paths []string, out io.Writer) {
	reader, cleanup, err := multiFile(paths)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
	extract(reader, out)
}

func getAllGoFiles(dir string) []string {
	var paths []string
	err := filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("walk %v: %v", dir, err)
	}
	return paths
}

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

func extract(r io.Reader, w io.Writer) {
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

	write := func(heading string, m map[string]string) {
		fmt.Fprintf(w, "# %s\n", heading)
		names := make([]string, 0, len(m))
		for k := range m {
			names = append(names, k)
		}
		sort.Strings(names)
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

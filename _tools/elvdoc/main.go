package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		extract(os.Stdin, os.Stdout)
	} else {
		for _, arg := range args {
			extractFile(arg, os.Stdout)
		}
	}
}

func extractFile(name string, w io.Writer) {
	f, err := os.Open(name)
	if err != nil {
		log.Printf("open %v: %v", name, err)
		return
	}
	defer f.Close()
	extract(f, w)
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
			varDocPrefix = "//elvish:doc-var "
			fnDocPrefix  = "//elvish:doc-fn "
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
				log.Printf("read: %v", err)
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

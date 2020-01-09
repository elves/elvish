package program

import (
	"fmt"
	"io"
	"os"
)

func usage(out io.Writer, f *flagSet) {
	fmt.Fprintln(out, "Usage: elvish [flags] [script]")
	fmt.Fprintln(out, "Supported flags:")
	f.PrintDefaults()
}

type helpProgram struct{ flag *flagSet }

func (p helpProgram) Main(fds [3]*os.File, _ []string) int {
	p.flag.SetOutput(fds[1])
	usage(fds[1], p.flag)
	return 0
}

type badUsageProgram struct {
	message string
	flag    *flagSet
}

func (p badUsageProgram) Main(fds [3]*os.File, _ []string) int {
	fds[2].WriteString(p.message + "\n")
	usage(fds[2], p.flag)
	return 2
}

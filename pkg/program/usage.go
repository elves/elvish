package program

import "os"

type helpProgram struct{ flag *flagSet }

func (p helpProgram) Main(fds [3]*os.File, _ []string) int {
	usage(fds[1], p.flag)
	return 0
}

type badUsageProgram struct {
	message string
	flag    *flagSet
}

func (p badUsageProgram) Main(fds [3]*os.File, _ []string) int {
	usage(fds[2], p.flag)
	return 2
}

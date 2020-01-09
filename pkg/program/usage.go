package program

import "os"

// ShowHelp shows help message.
type ShowHelp struct {
	flag *flagSet
}

func (s ShowHelp) Main(fds [3]*os.File, _ []string) int {
	usage(fds[1], s.flag)
	return 0
}

type ShowCorrectUsage struct {
	message string
	flag    *flagSet
}

func (s ShowCorrectUsage) Main(fds [3]*os.File, _ []string) int {
	usage(fds[2], s.flag)
	return 2
}

package sys

import (
	"github.com/mattn/go-isatty"
)

func IsATTY(fd int) bool {
	return isatty.IsTerminal(uintptr(fd)) ||
		isatty.IsCygwinTerminal(uintptr(fd))
}

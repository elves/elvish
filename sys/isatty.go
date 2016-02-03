package sys

/*
#include <unistd.h>
*/
import "C"

func IsATTY(fd int) bool {
	return C.isatty(C.int(fd)) != 0
}

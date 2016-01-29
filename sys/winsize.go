package sys

/*
#include <termios.h>
#include <sys/ioctl.h>

void getwinsize(int fd, int *row, int *col) {
	struct winsize wsz;
	ioctl(fd, TIOCGWINSZ, &wsz);
	*row = wsz.ws_row;
	*col = wsz.ws_col;
}
*/
import "C"

// GetWinsize queries the size of the terminal referenced by the given file
// descriptor.
func GetWinsize(fd int) (row, col int) {
	var r, c C.int
	C.getwinsize(C.int(fd), &r, &c)
	return int(r), int(c)
}

package sys

/*
#include <unistd.h>

int f(int fd, pid_t pid) {
	return tcsetpgrp(fd, pid);
}
*/
import "C"
import "syscall"

func Tcsetpgrp(fd int, pid int) error {
	i := syscall.Errno(C.f(C.int(fd), C.pid_t(pid)))
	if i != 0 {
		return syscall.Errno(i)
	}
	return nil
}

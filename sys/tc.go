package sys

/*
#include <unistd.h>
#include <errno.h>

int f(int fd, pid_t pid) {
	return tcsetpgrp(fd, pid);
}

int e() {
	return errno;
}
*/
import "C"
import "syscall"

func Tcsetpgrp(fd int, pid int) error {
	i := C.f(C.int(fd), C.pid_t(pid))
	if i != 0 {
		return syscall.Errno(C.e())
	}
	return nil
}

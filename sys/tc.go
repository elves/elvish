package sys

/*
#include <unistd.h>
#include <errno.h>

pid_t get(int fd) {
	return tcgetpgrp(fd);
}

int set(int fd, pid_t pid) {
	return tcsetpgrp(fd, pid);
}

int e() {
	return errno;
}
*/
import "C"
import "syscall"

func Tcgetpgrp(fd int) (int, error) {
	i := C.get(C.int(fd))
	if i == -1 {
		return -1, syscall.Errno(C.e())
	}
	return int(i), nil
}

func Tcsetpgrp(fd int, pid int) error {
	i := C.set(C.int(fd), C.pid_t(pid))
	if i != 0 {
		return syscall.Errno(C.e())
	}
	return nil
}

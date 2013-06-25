#include <stdarg.h>
#include <unistd.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>

#include "common.h"

pid_t root_pid;

static void child_header() {
    pid_t pid = getpid();
    if (pid != root_pid) {
        fprintf(stderr, "(child %d) ", pid);
    }
}

void Say(const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);

    child_header();
    vfprintf(stderr, fmt, ap);

    va_end(ap);
}

void DieIf(int cond, const char *s) {
    if (cond) {
        child_header();
        perror(s);
        exit(1);
    }
}

void DieIf_1(int ret, const char *s) {
    DieIf(ret == -1, s);
}

char *Itos(int i) {
    char *buf = 0;
    int n = 2;
    while (1) {
        buf = realloc(buf, n);
        if (snprintf(buf, n, "%d", i) < n) {
            return buf;
        } else {
            n *= 2;
        }
    }
}

void SetCloexec(int fd) {
    int f = fcntl(fd, F_GETFD);
    DieIf_1(f, "fcntl");
    DieIf_1(fcntl(fd, F_SETFD, f | FD_CLOEXEC), "fcntl");
}

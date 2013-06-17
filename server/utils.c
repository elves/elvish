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

void say(const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);

    child_header();
    vfprintf(stderr, fmt, ap);

    va_end(ap);
}

void check_1(const char *s, int ret) {
    if (ret == -1) {
        child_header();
        perror(s);
        exit(1);
    }
}

char *slurp(int fd) {
    int cap = 32;
    int begin = 0;
    char *buf = malloc(cap);
    while (1) {
        if (cap == begin) {
            cap *= 2;
            buf = realloc(buf, cap);
        }
        int nr = read(fd, buf + begin, cap - begin);
        if (nr < 0) {
            free(buf);
            return 0;
        } else if (nr == 0) {
            buf[begin] = '\0';
            return buf;
        } else {
            begin += nr;
        }
    }
}

char *itos(int i) {
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

void set_cloexec(int fd) {
    int f = fcntl(fd, F_GETFD);
    check_1("fcntl", f);
    check_1("fcntl", fcntl(fd, F_SETFD, f | FD_CLOEXEC));
}

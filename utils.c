#include <stdarg.h>
#include <unistd.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>

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

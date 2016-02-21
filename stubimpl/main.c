#include <stdio.h>
#include <stdlib.h>
#include <signal.h>
#include <unistd.h>
#include <string.h>

void must(int ok, char *s) {
    if (!ok) {
        perror(s);
        exit(1);
    }
}

void handler(int signum) {
    char s[5] = "   \n";
    char *p = &s[3];
    while (signum > 0) {
        p--;
        *p = (signum % 10) + '0';
        signum /= 10;
    }
    must(write(1, p, s+4-p) != -1, "write signum");
}

enum { ARGV0BUF = 32 };

char argv0buf[ARGV0BUF];

int main(int argc, char **argv) {
    int i;
    for (i = 1; i <= 64; i++) {
        signal(i, handler);
    }
    signal(SIGTTIN, SIG_IGN);
    signal(SIGTTOU, SIG_IGN);

    must(write(1, "ok\n", 3) != -1, "write ok");

    while (1) {
        if (fgets(argv0buf, ARGV0BUF, stdin) == NULL) {
            if (feof(stdin)) {
                exit(0);
            } else {
                exit(10);
            }
        }
        int n = strlen(argv0buf);
        if (n > 0 && argv0buf[n-1] == '\n') {
            argv0buf[n-1] = '\0';
        }
        strcpy(argv[0], argv0buf);
    }
}

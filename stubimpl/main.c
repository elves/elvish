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

enum { ARGV0MAX = 32 };

int main(int argc, char **argv) {
    int i;
    for (i = 1; i <= 64; i++) {
        signal(i, handler);
    }
    signal(SIGTTIN, SIG_IGN);
    signal(SIGTTOU, SIG_IGN);

    must(write(1, "ok\n", 3) != -1, "write ok");

    char op;
    int len;
    char *buf;
    int scanned;
    while ((scanned = scanf(" %c%d ", &op, &len)) == 2) {
        // printf("op=%d, len=%d\n", op, len);
        buf = malloc(len+1);
        int nr = read(0, buf, len);
        buf[nr] = '\0';
        // printf("buf=%s\n", buf);
        if (op == 'd') {
            // Change directory.
            chdir(buf);
        } else if (op == 't') {
            if (len > ARGV0MAX) {
                buf[ARGV0MAX] = '\0';
            }
            strcpy(argv[0], buf);
        }
        free(buf);
    }
    if (scanned != EOF) {
        must(write(1, "bad msg\n", 8), "write bad msg");
        while (getchar() != EOF)
            ;
        return 1;
    }
    return 0;
}

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

void badmsg() {
    must(write(1, "bad msg\n", 8), "write bad msg");
}

// read as much as possible, but no more than n.
int readn(int fd, char *buf, int n) {
    int nr, nrall;
    nrall = 0;
    while (n > 0 && (nr = read(fd, buf, n)) > 0) {
        buf += nr;
        nrall += nr;
        n -= nr;
    }
    return nrall;
}

int main(int argc, char **argv) {
    int i;
    for (i = 1; i <= 64; i++) {
        signal(i, handler);
    }
    signal(SIGTTIN, SIG_IGN);
    signal(SIGTTOU, SIG_IGN);

    must(write(1, "ok\n", 3) != -1, "write ok");

    int nr;
    char opbuf[6] = "12345";
    while ((nr = readn(0, opbuf, 5)) == 5) {
        char opcode = opbuf[0];
        int len = atoi(opbuf+1);
        char *buf = malloc(len+1);
        buf[len] = '\0';
        fprintf(stderr, "code = %c, len = %d\n", opcode, len);
        if (readn(0, buf, len) < len) {
            free(buf);
            break;
        }
        fprintf(stderr, "data = %s\n", buf);
        switch (opcode) {
        case 'd':
            // Change directory.
            chdir(buf);
            break;
        case 't':
            if (len > ARGV0MAX) {
                buf[ARGV0MAX] = '\0';
            }
            strcpy(argv[0], buf);
            break;
        default:
            badmsg();
        }
        free(buf);
    }

    if (nr != 0) {
        badmsg();
    }

    return 0;
}

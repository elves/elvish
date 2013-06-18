#include <stdio.h>
#include <jansson.h>

#include "common.h"
#include "res.h"

FILE *resFile;

ResCmd *NewResCmd() {
    // XXX STUB
    return 0;
}

ResProcState *NewResProcState() {
    // XXX STUB
    return 0;
}

void FreeRes(Res *r) {
    // XXX STUB
}

int WriteRes(const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    int r = vfprintf(resFile, fmt, ap);
    va_end(ap);
    return r;
}

int SendRes(Res *r) {
    // XXX STUB
    return 0;
}

void InitRes(int fd) {
    SetCloexec(fd);
    resFile = fdopen(fd, "w");
    setlinebuf(resFile);
}

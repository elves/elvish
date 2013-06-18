#include <stdio.h>
#include <jansson.h>

#include "common.h"
#include "res.h"

FILE *resFile;

ResCmd *NewResCmd() {
    ResCmd *r = alloc(ResCmd, 1);
    r->type = RES_TYPE_CMD;
    return r;
}

ResProcState *NewResProcState() {
    ResProcState *r = alloc(ResProcState, 1);
    r->type = RES_TYPE_PROC_STATE;
    return r;
}

ResBadRequest *NewResBadRequest() {
    ResBadRequest *r = alloc(ResBadRequest, 1);
    r->type = RES_TYPE_BAD_REQUEST;
    return r;
}

void freeResBadRequest(ResBadRequest *r) {
    if (r->err) {
        free(r->err);
    }
    free(r);
}

void FreeRes(Res *r) {
    switch (r->type) {
    case RES_TYPE_BAD_REQUEST:
        freeResBadRequest((ResBadRequest*)r);
        break;
    default:
        free(r);
    }
}

int WriteRes(const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    int r = vfprintf(resFile, fmt, ap);
    va_end(ap);
    return r;
}

json_t *buildResCmd(ResCmd *r) {
    return json_pack("{si}", "Pid", r->pid);
}

json_t *buildResProcState(ResProcState *r) {
    return json_pack("{si sb si sb si sb sb si sb}",
                     "Pid", r->pid,
                     "Exited", r->exited,
                     "ExitStatus", r->exitStatus,
                     "Signaled", r->signaled,
                     "TermSig", r->termSig,
                     "CoreDump", r->coreDump,
                     "Stopped", r->stopped,
                     "StopSig", r->stopSig,
                     "Continued", r->continued);
}

json_t *buildResBadRequest(ResBadRequest *r) {
    return json_pack("{si}", "Err", r->err);
}

int SendRes(Res *r) {
    const char *type;
    json_t *data;
    if (r->type == RES_TYPE_CMD) {
        type = "cmd";
        data = buildResCmd((ResCmd*)r);
    } else if (r->type == RES_TYPE_PROC_STATE) {
        type = "procState";
        data = buildResProcState((ResProcState*)r);
    } else if (r->type == RES_TYPE_BAD_REQUEST) {
        type = "badRequest";
        data = buildResBadRequest((ResBadRequest*)r);
    } else {
        return -1;
    }
    json_t *root = json_pack("{ss so}", "Type", type, "Data", data);
    json_dumpf(root, resFile, JSON_COMPACT); // XXX check return value
    fprintf(resFile, "\n");
    json_decref(root);
    return 0;
}

void InitRes(int fd) {
    SetCloexec(fd);
    resFile = fdopen(fd, "w");
    setlinebuf(resFile);
}

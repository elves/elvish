#ifndef _res_h_
#define _res_h_

#include <stdbool.h>
#include <unistd.h>

typedef enum {
    RES_TYPE_BAD_REQUEST,
    RES_TYPE_CMD,
    RES_TYPE_PROC_STATE,
} ResType;

#define RES_HEADER ResType type

typedef struct {
    RES_HEADER;
} Res;

typedef struct {
    RES_HEADER;
    char *err;
} ResBadRequest;

typedef struct {
    RES_HEADER;
    pid_t pid;
} ResCmd;

typedef struct {
    RES_HEADER;
    pid_t pid;
    bool exited;
    int exitStatus;
    bool signaled;
    int termSig;
    bool coreDump;
    bool stopped;
    int stopSig;
    bool continued;
} ResProcState;

ResCmd *NewResCmd();
ResProcState *NewResProcState();
ResBadRequest *NewResBadRequest();

void FreeRes(Res *r);
int SendRes(Res *r);
void InitRes(int fd);

#endif

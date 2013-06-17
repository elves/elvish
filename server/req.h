#ifndef __REQ_H
#define __REQ_H

#include <jansson.h>

typedef enum {
    REQ_TYPE_COMMAND,
    REQ_TYPE_EXIT,
} ReqType;

#define REQ_HEADER ReqType type

typedef struct {
    REQ_HEADER;
} Req;

typedef struct {
    REQ_HEADER;
} ReqExit;

typedef struct {
    REQ_HEADER;
    char *path;
    char **argv;
    char **envp;
} ReqCmd;

void FreeReq(Req *r);
Req *RecvReq(char **err);
void InitReq(int fd);

#endif

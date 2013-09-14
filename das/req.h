#ifndef _req_h_
#define _req_h_

typedef enum {
    REQ_TYPE_CMD,
    REQ_TYPE_EXIT,
} ReqType;

enum {
    FD_CLOSE = -1,
    FD_SEND = -2,
};

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
    int (*redirs)[2];
    bool *isRecvedFd;
} ReqCmd;

void FreeReq(Req *r);
Req *RecvReq(char **err);

#endif

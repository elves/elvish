#ifndef __REQ_H
#define __REQ_H

#include <jansson.h>

typedef enum {
    REQ_TYPE_COMMAND,
    REQ_TYPE_EXIT,
} req_type_t;

#define REQ_HEADER req_type_t type

typedef struct {
    REQ_HEADER;
} req_t;

typedef struct {
    REQ_HEADER;
} req_exit_t;

typedef struct {
    REQ_HEADER;
    char *path;
    char **argv;
    char **envp;
} req_command_t;

void free_req(req_t *r);
char *recv_req(req_t **r);
void init_req(int fd);

#endif

#ifndef __REQ_H
#define __REQ_H

#include <jansson.h>

typedef struct {
    char *path;
    char **argv;
    char **envp;
} command_t;

void free_command(command_t *p);
char *recv_req(command_t **cmd);
void init_req(int fd);

#endif

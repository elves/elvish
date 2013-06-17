#ifndef __REQ_H
#define __REQ_H

#include <jansson.h>

typedef struct {
    char *path;
    char **argv;
    char **envp;
} command_t;

void free_strings(char **p);
void free_command(command_t *p);
void print_command(command_t *cmd);
command_t *parse_command(json_t *root);

#endif

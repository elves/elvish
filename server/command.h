#ifndef __COMMAND_H
#define __COMMAND_H

typedef struct {
    char *path;
    char **argv;
    char **envp;
} command_t;

#endif

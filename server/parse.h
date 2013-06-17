#ifndef __PARSE_H
#define __PARSE_H

#include <jansson.h>

#include "command.h"

void free_strings(char **p);
void free_command(command_t *p);
void print_command(command_t *cmd);
command_t *parse_command(json_t *root);

#endif

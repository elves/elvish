#ifndef _common_h_
#define _common_h_

#include <stdbool.h>
#include <sys/types.h>

#define alloc(type, n) ((type*) calloc(sizeof(type), (n)))

extern pid_t root_pid;

void Say(const char *fmt, ...);
void DieIf(int cond, const char *s);
void DieIf_1(int ret, const char *s);
char *Itos(int i);
void SetCloexec(int fd);

#endif

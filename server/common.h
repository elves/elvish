#include <sys/types.h>

#define alloc(type, n) ((type*) calloc(sizeof(type), (n)))

extern pid_t root_pid;

void Say(const char *fmt, ...);
void Check_1(const char *s, int ret);
char *Itos(int i);
void SetCloexec(int fd);

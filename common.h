#include <sys/types.h>

#define alloc(type, n) ((type*) calloc(sizeof(type), (n)))

extern pid_t root_pid;

void say(const char *fmt, ...);
void check_1(const char *s, int ret);
char *slurp(int fd);
char *itos(int i);

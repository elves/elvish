#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <sys/wait.h>
#include <sys/socket.h>
#include <sys/un.h>

#include "common.h"
#include "req.h"

extern char **environ;

FILE *res;
int exiting = 0;

void external(command_t *cmd) {
    environ = cmd->envp;
    check_1("exec", execv(cmd->path, cmd->argv));
}

void worker() {
    json_t *root;
    json_error_t error;

    char *buf = recv_req();
    if (!buf) {
        exiting = 1;
        return;
    }
    root = json_loads(buf, 0, &error);
    free(buf);
    if (!root) {
        fprintf(res, "json: error on line %d: %s\n", error.line, error.text);
        return;
    }

    command_t *cmd = parse_command(root);

    if (!cmd) {
        fprintf(res, "json: command doesn't conform to schema\n");
        return;
    }

    pid_t pid;
    check_1("fork", pid = fork());
    if (pid == 0) {
        fclose(res);
        external(cmd);
    } else {
        fprintf(res, "spawned external: pid = %d\n", pid);
        while (1) {
            int status;
            pid_t ret = waitpid(pid, &status, 0);
            if (ret == -1 && errno == ECHILD) {
                break;
            }
            check_1("wait", ret);
            printf("external %d ", pid);
            if (WIFEXITED(status)) {
                printf("terminated: %d\n", WEXITSTATUS(status));
            } else if (WIFSIGNALED(status)) {
                printf("terminated by signal: %d\n", WTERMSIG(status));
            } else if (WCOREDUMP(status)) {
                printf("core dumped\n");
            } else if (WIFSTOPPED(status)) {
                printf("stopped by signal: %d\n", WSTOPSIG(status));
            } else if (WIFCONTINUED(status)) {
                printf("continued\n");
            } else {
                printf("changed to some state das doesn't know\n");
            }
        }
    }
    json_decref(root);
    free_command(cmd);
}

int main(int argc, char **argv) {
    if (argc > 2) {
        fprintf(stderr, "Usage: das [path to dasc]\n");
        return 1;
    }

    root_pid = getpid();

    int reqp[2], resp[2];
    pipe(reqp);
    pipe(resp);

    pid_t pid;
    check_1("fork", pid = fork());
    if (pid == 0) {
        // Child: write to req, read from res
        close(reqp[0]);
        close(resp[1]);

        // exec dasc
        char *path;
        if (argc == 2 && argv[1][0] == '/') {
            path = argv[1];
        } else {
            const char *relpath = argc == 2 ? argv[1] : "dasc";
            int nrel = strlen(relpath);
            int n = 256;
            char *buf = 0;
            while (1) {
                buf = realloc(buf, n + nrel + 1);
                if (getcwd(buf, n)) {
                    break;
                } else if (errno != ERANGE) {
                    check_1("getcwd", -1);
                }
                n *= 2;
            }
            path = buf;
            strcat(path, "/");
            strcat(path, relpath);
        }
        check_1("exec", execl(path, path, itos(reqp[1]), itos(resp[0]), 0));
    }

    // Parent: read from req, write to res
    close(reqp[1]);
    close(resp[0]);
    init_req(reqp[0]);
    res = fdopen(resp[1], "w");
    setlinebuf(res);

    do {
        worker();
    } while (!exiting);

    int status;
    do {
        check_1("wait", waitpid(pid, &status, 0));
    } while (!WIFEXITED(status) && !WIFSIGNALED(status));

    return 0;
}

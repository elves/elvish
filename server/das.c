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
#include "res.h"

extern char **environ;

int exiting = 0;

void external(ReqCmd *cmd) {
    environ = cmd->envp;
    Check_1("exec", execv(cmd->path, cmd->argv));
}

void worker() {
    char *err;
    Req *req = RecvReq(&err);
    if (!req) {
        WriteRes("%s\n", err);
        return;
    }

    ReqType type = req->type;
    if (type == REQ_TYPE_CMD) {
        pid_t pid;
        Check_1("fork", pid = fork());
        if (pid == 0) {
            external((ReqCmd*)req);
        } else {
            ResCmd *res = NewResCmd();
            res->pid = pid;
            SendRes((Res*)res);
            free(res);
            while (1) {
                int status;
                pid_t ret = waitpid(pid, &status, 0);
                if (ret == -1 && errno == ECHILD) {
                    break;
                }
                Check_1("wait", ret);

                ResProcState *res = NewResProcState();
                res->pid = pid;
                res->exited = WIFEXITED(status);
                if (res->exited) {
                    res->exitStatus = WEXITSTATUS(status);
                }
                res->signaled = WIFSIGNALED(status);
                if (res->signaled) {
                    res->termSig = WTERMSIG(status);
                }
                res->coreDump = WCOREDUMP(status);
                res->stopped = WIFSTOPPED(status);
                if (res->stopped) {
                    res->stopSig = WSTOPSIG(status);
                }
                res->continued = WIFCONTINUED(status);
                SendRes((Res*)res);
                free(res);
            }
        }
    } else if (type == REQ_TYPE_EXIT) {
        exiting = 1;
    }

    FreeReq(req);
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
    Check_1("fork", pid = fork());
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
                    Check_1("getcwd", -1);
                }
                n *= 2;
            }
            path = buf;
            strcat(path, "/");
            strcat(path, relpath);
        }
        Check_1("exec", execl(path, path, Itos(reqp[1]), Itos(resp[0]), 0));
    }

    // Parent: read from req, write to res
    close(reqp[1]);
    close(resp[0]);
    InitReq(reqp[0]);
    InitRes(resp[1]);

    do {
        worker();
    } while (!exiting);

    int status;
    do {
        Check_1("wait", waitpid(pid, &status, 0));
    } while (!WIFEXITED(status) && !WIFSIGNALED(status));

    return 0;
}

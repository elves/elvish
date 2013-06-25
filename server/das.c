#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <sys/wait.h>
#include <sys/socket.h>
#include <sys/un.h>

#include "common.h"
#include "tube.h"
#include "req.h"
#include "res.h"

extern char **environ;

int exiting = 0;

void external(ReqCmd *cmd) {
    environ = cmd->envp;
    if (cmd->redirOutput) {
        DieIf_1(dup2(cmd->output, 1), "dup2");
        close(cmd->output);
    }
    DieIf_1(execv(cmd->path, cmd->argv), "exec");
}

void worker() {
    char *err;
    Req *req = RecvReq(&err);
    if (!req) {
        ResBadRequest *res = NewResBadRequest();
        res->err = err;
        SendRes((Res*)res);
        FreeRes((Res*)res);
        return;
    }

    ReqType type = req->type;
    if (type == REQ_TYPE_CMD) {
        ReqCmd *cmd = (ReqCmd*)req;
        pid_t pid;
        DieIf_1(pid = fork(), "fork");
        if (pid == 0) {
            external(cmd);
        } else {
            if (cmd->redirOutput) {
                close(cmd->output);
            }
            ResCmd *res = NewResCmd();
            res->pid = pid;
            SendRes((Res*)res);
            FreeRes((Res*)res);
            while (1) {
                int status;
                pid_t ret = waitpid(pid, &status, 0);
                if (ret == -1 && errno == ECHILD) {
                    break;
                }
                DieIf_1(ret, "wait");

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
                FreeRes((Res*)res);
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

    int textTube[2];
    int fdTube[2];
    DieIf_1(socketpair(AF_UNIX, SOCK_STREAM, 0, textTube), "socketpair");
    DieIf_1(socketpair(AF_UNIX, SOCK_STREAM, 0, fdTube), "socketpair");

    pid_t pid;
    DieIf_1(pid = fork(), "fork");
    if (pid == 0) {
        // Child uses *Tube[0] - may result in smaller fd :)
        close(textTube[1]);
        close(fdTube[1]);

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
                    DieIf(1, "getcwd");
                }
                n *= 2;
            }
            path = buf;
            strcat(path, "/");
            strcat(path, relpath);
        }
        DieIf_1(execl(path, path, Itos(textTube[0]), Itos(fdTube[0]), 0),
                "exec");
    }

    // Parent uses *Tube[1]
    close(textTube[0]);
    close(fdTube[0]);
    InitTubes(textTube[1], fdTube[1]);

    do {
        worker();
    } while (!exiting);

    int status;
    do {
        DieIf_1(waitpid(pid, &status, 0), "wait");
    } while (!WIFEXITED(status) && !WIFSIGNALED(status));

    return 0;
}

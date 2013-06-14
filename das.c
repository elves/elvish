#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <sys/wait.h>
#include <sys/socket.h>
#include <sys/un.h>

#include "common.h"
#include "command.h"
#include "parse.h"

extern char **environ;

int make_server_socket(const char *path) {
    int listener;
    check_1("socket", listener = socket(AF_UNIX, SOCK_STREAM, 0));

    struct sockaddr_un local = {AF_UNIX};
    strncpy(local.sun_path, path, sizeof(local.sun_path));
    local.sun_path[sizeof(local.sun_path) - 1] = '\0';
    path = local.sun_path;
    if (access(path, F_OK) != -1) {
        check_1("unlink", unlink(local.sun_path));
    }

    check_1("bind",
            bind(listener, (struct sockaddr*) &local,
                 strlen(path) + sizeof(local.sun_family)));

    check_1("listen", listen(listener, 128));

    return listener;
}

void external(command_t *cmd) {
    environ = cmd->envp;
    check_1("exec", execv(cmd->path, cmd->argv));
}

void worker(int socket) {
    json_t *root;
    json_error_t error;

    char *buf = slurp(socket);
    close(socket);
    root = json_loads(buf, 0, &error);
    free(buf);
    if (!root) {
        say("json: error on line %d: %s\n", error.line, error.text);
        exit(1);
    }

    command_t *cmd = parse_command(root);

    if (!cmd) {
        say("json: command doesn't conform to schema\n");
        exit(1);
    }

    pid_t pid;
    check_1("fork", pid = fork());
    if (pid == 0) {
        external(cmd);
    } else {
        printf("spawned external: pid = %d\n", pid);
        while (1) {
            int status;
            pid = wait(&status);
            if (pid == -1 && errno == ECHILD) {
                break;
            }
            check_1("wait", pid);
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
                printf("changed to some state dasd doesn't know\n");
            }
        }
    }
    json_decref(root);
    free_command(cmd);
    exit(0);
}

int main() {
    root_pid = getpid();
    int listener = make_server_socket("/tmp/das");

    while (1) {
        struct sockaddr_un remote;
        unsigned len = sizeof(remote);
        int socket;
        check_1("accept",
                socket = accept(listener, (struct sockaddr*)&remote, &len));

        say("accepted a request\n");

        pid_t pid;
        check_1("fork", pid = fork());
        if (pid == 0) {
            worker(socket);
        } else {
            close(socket);
            say("spawned worker: pid = %d, socket = %d\n", pid, socket);
        }
    }

    close(listener);
    return 0;
}

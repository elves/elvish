#define _GNU_SOURCE
#include <stdio.h>
#include <fcntl.h>
#include <stdlib.h>
#include <string.h>

#include <jansson.h>

#include "common.h"
#include "req.h"

FILE *g_reqfile;

void freeStrings(char **p) {
    char **q;
    for (q = p; *q; q++) {
        free(*q);
    }
    free(p);
}

void freeReqCmd(ReqCmd *p) {
    if (p->path) {
        free(p->path);
    }
    if (p->argv) {
        freeStrings(p->argv);
    }
    if (p->envp) {
        freeStrings(p->envp);
    }
    free(p);
}

void freeReqExit(ReqExit *p) {
    free(p);
}

void FreeReq(Req *p) {
    switch (p->type) {
    case REQ_TYPE_COMMAND:
        freeReqCmd((ReqCmd*)p);
        break;
    case REQ_TYPE_EXIT:
        freeReqExit((ReqExit*)p);
        break;
    }
}

void printReqCmd(ReqCmd *cmd) {
    char **a;
    printf("path: %s\n", cmd->path);
    printf("args:\n");
    for (a = cmd->argv; *a; a++) {
        printf("      %s\n", *a);
    }
}

char *loadString(json_t *root) {
    if (!json_is_string(root)) {
        fprintf(stderr, "string\n");
        return 0;
    }
    return strdup(json_string_value(root));
}

char **loadArgv(json_t *root) {
    if (!json_is_array(root)) {
        fprintf(stderr, "argv not array\n");
        return 0;
    }

    int n = json_array_size(root);
    char **argv = alloc(char*, n + 1);

    int i;
    for (i = 0; i < n; i++) {
        json_t *arg = json_array_get(root, i);
        if (!json_is_string(arg)) {
            fprintf(stderr, "argv element not string\n");
            freeStrings(argv);
            return 0;
        }
        argv[i] = strdup(json_string_value(arg));
    }
    return argv;
}

char **loadEnvp(json_t *root) {
    if (!json_is_object(root)) {
        fprintf(stderr, "envp not object\n");
        return 0;
    }

    int n = json_object_size(root);
    char **envp = alloc(char*, n + 1);

    const char *key;
    json_t *value;
    int i = 0;
    json_object_foreach(root, key, value) {
        if (!json_is_string(value)) {
            fprintf(stderr, "envp value not object\n");
            freeStrings(envp);
            return 0;
        }
        const char *value_s = json_string_value(value);
        envp[i] = (char*)malloc(strlen(key) + strlen(value_s) + 2);
        strcpy(envp[i], key);
        strcat(envp[i], "=");
        strcat(envp[i], value_s);
        i++;
    }
    return envp;
}

ReqCmd *newReqCmd() {
    ReqCmd *r = alloc(ReqCmd, 1);
    r->type = REQ_TYPE_COMMAND;
    return r;
}

ReqCmd *loadReqCmd(json_t *root) {
    ReqCmd *cmd = newReqCmd();

    const char *path;
    json_t *args, *env;
    int success =
        (!json_unpack_ex(root, 0, JSON_STRICT, "{ss so so}",
                         "path", &path, "args", &args, "env", &env) &&
         (cmd->argv = loadArgv(args)) &&
         (cmd->envp = loadEnvp(env)));

    if (success) {
        cmd->path = strdup(path);
        return cmd;
    } else {
        freeReqCmd(cmd);
        return 0;
    }
}

ReqExit *newReqExit() {
    ReqExit *r = alloc(ReqExit, 1);
    r->type = REQ_TYPE_EXIT;
    return r;
}

Req *loadReq(json_t *root) {
    char *type;
    json_t *data;
    if (json_unpack_ex(root, 0, JSON_STRICT, "{ss so}",
                       "type", &type, "data", &data)) {
        return 0;
    }
    if (!strcmp(type, "command")) {
        return (Req*)loadReqCmd(data);
    } else {
        // TODO error("bad request type")
        return 0;
    }
}

char *readReq() {
    char *buf = 0;
    size_t n;
    if (getline(&buf, &n, g_reqfile) == -1) {
        return 0;
    }
    return buf;
}

Req *RecvReq(char **err) {
    char *buf = readReq();
    if (!buf) {
        return (Req*)newReqExit();
    }

    json_t *root;
    json_error_t error;
    root = json_loads(buf, 0, &error);
    free(buf);

    if (!root) {
        asprintf(err, "json: error on line %d: %s", error.line, error.text);
        return 0;
    }

    Req *cmd = loadReq(root);
    json_decref(root);

    if (!cmd) {
        *err = strdup("json: command doesn't conform to schema");
    }

    return cmd;
}

void InitReq(int fd) {
    set_cloexec(fd);
    g_reqfile = fdopen(fd, "r");
}

#define _GNU_SOURCE
#include <stdio.h>
#include <fcntl.h>
#include <stdlib.h>
#include <string.h>

#include <jansson.h>

#include "common.h"
#include "req.h"

FILE *req;

void free_strings(char **p) {
    char **q;
    for (q = p; *q; q++) {
        free(*q);
    }
    free(p);
}

void free_command(req_command_t *p) {
    if (p->path) {
        free(p->path);
    }
    if (p->argv) {
        free_strings(p->argv);
    }
    if (p->envp) {
        free_strings(p->envp);
    }
    free(p);
}

void free_req(req_t *p) {
    free_command((req_command_t*)p);
}

void print_command(req_command_t *cmd) {
    char **a;
    printf("path: %s\n", cmd->path);
    printf("args:\n");
    for (a = cmd->argv; *a; a++) {
        printf("      %s\n", *a);
    }
}

char *parse_string(json_t *root) {
    if (!json_is_string(root)) {
        fprintf(stderr, "string\n");
        return 0;
    }
    return strdup(json_string_value(root));
}

char **parse_argv(json_t *root) {
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
            free_strings(argv);
            return 0;
        }
        argv[i] = strdup(json_string_value(arg));
    }
    return argv;
}

char **parse_envp(json_t *root) {
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
            free_strings(envp);
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

req_command_t *parse_command(json_t *root) {
    req_command_t *cmd = alloc(req_command_t, 1);

    const char *path;
    json_t *args, *env;
    int success =
        (!json_unpack_ex(root, 0, JSON_STRICT, "{ss so so}",
                         "path", &path, "args", &args, "env", &env) &&
         (cmd->argv = parse_argv(args)) &&
         (cmd->envp = parse_envp(env)));

    if (success) {
        cmd->path = strdup(path);
        return cmd;
    } else {
        free_command(cmd);
        return 0;
    }
}

char *read_req() {
    char *buf = 0;
    size_t n;
    if (getline(&buf, &n, req) == -1) {
        return 0;
    }
    return buf;
}

extern int exiting;

char *recv_req(req_t **cmd) {
    char *buf = read_req();
    if (!buf) {
        exiting = 1;
        return strdup("exiting");
    }

    json_t *root;
    json_error_t error;
    root = json_loads(buf, 0, &error);
    free(buf);

    if (!root) {
        char *err;
        asprintf(&err, "json: error on line %d: %s", error.line, error.text);
        return err;
    }

    *cmd = (req_t*)parse_command(root);
    json_decref(root);

    if (!*cmd) {
        return strdup("json: command doesn't conform to schema");
    }

    return 0;
}

void init_req(int fd) {
    set_cloexec(fd);
    req = fdopen(fd, "r");
}

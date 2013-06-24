#include "common.h"
#include "tube.h"

FILE *TubeFile;

void InitTube(int fd) {
    TubeFile = fdopen(fd, "a+");
    if (!TubeFile) {
        Check_1("fdopen", -1);
    }
}

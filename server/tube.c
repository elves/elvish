#include "common.h"
#include "tube.h"

int FdTube;
FILE *TextTube;

void InitTubes(int textTube, int fdTube) {
    SetCloexec(textTube);
    SetCloexec(fdTube);
    TextTube = fdopen(textTube, "a+");
    if (!TextTube) {
        DieIf_1(-1, "fdopen");
    }
    FdTube = fdTube;
}

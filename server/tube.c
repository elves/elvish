#include "common.h"
#include "tube.h"

int FdTube;
FILE *TextTube;

void InitTubes(int textTube, int fdTube) {
    SetCloexec(textTube);
    SetCloexec(fdTube);
    TextTube = fdopen(textTube, "a+");
    DieIf(TextTube == 0, "fdopen");
    FdTube = fdTube;
}

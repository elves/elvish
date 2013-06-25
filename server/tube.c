#include "common.h"
#include "tube.h"

int FdTube;
FILE *TextTube;

void InitTubes(int textTube, int fdTube) {
    SetCloexec(textTube);
    SetCloexec(fdTube);
    TextTube = fdopen(textTube, "a+");
    if (!TextTube) {
        Check_1("fdopen", -1);
    }
    FdTube = fdTube;
}

#include "common.h"
#include "tube.h"

int FdTubeFd;
FILE *TubeFile;

void InitTubes(int textTube, int fdTube) {
    TubeFile = fdopen(textTube, "a+");
    if (!TubeFile) {
        Check_1("fdopen", -1);
    }
    FdTubeFd = fdTube;
}

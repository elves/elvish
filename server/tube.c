#include "common.h"
#include "tube.h"

int TubeFd;
int FdTubeFd;
FILE *TubeFile;

void InitTubes(int textTube, int fdTube) {
    TubeFd = textTube;
    TubeFile = fdopen(textTube, "a+");
    if (!TubeFile) {
        Check_1("fdopen", -1);
    }
    FdTubeFd = fdTube;
}

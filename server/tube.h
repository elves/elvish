#ifndef _tube_h_
#define _tube_h_

#include <stdio.h>

extern int TubeFd;
extern FILE *TextTube;
extern int FdTube;

void InitTubes(int textTube, int fdTube);

#endif

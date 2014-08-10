package persistent

const (
	bitChunk   = 5
	nodeCap    = 1 << bitChunk
	tailMaxLen = nodeCap
	mask       = nodeCap - 1
)

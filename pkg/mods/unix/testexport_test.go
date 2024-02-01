//go:build unix

package unix

type RlimT = rlimT

var (
	GetRlimit = &getRlimit
	SetRlimit = &setRlimit
)

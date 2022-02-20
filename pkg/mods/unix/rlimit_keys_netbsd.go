package unix

func init() {
	// https://man.netbsd.org/getrlimit.2
	//
	// RLIMIT_NTHR and RLIMIT_SBSIZE are missing from x/sys/unix; the values are
	// taken from https://github.com/NetBSD/src/blob/trunk/sys/sys/resource.h.
	addRlimitKeys(map[int]string{
		11: "nthr",
		9:  "sbsize",
	})
}

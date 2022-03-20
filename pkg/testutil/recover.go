package testutil

func Recover(f func()) (r any) {
	defer func() { r = recover() }()
	f()
	return
}

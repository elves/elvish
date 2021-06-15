package testutil

func Recover(f func()) (r interface{}) {
	defer func() { r = recover() }()
	f()
	return
}

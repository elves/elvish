package testutil

// Umask sets the umask for the duration of the test, and restores it afterwards.
func Umask(c Cleanuper, m int) {
	save := umask(m)
	c.Cleanup(func() { _ = umask(save) })
}

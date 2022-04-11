package testutil

func Set[T any](c Cleanuper, p *T, v T) {
	old := *p
	*p = v
	c.Cleanup(func() { *p = old })
}

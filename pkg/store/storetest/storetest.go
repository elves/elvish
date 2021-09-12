// Package storetest keeps test suites against storedefs.Store.
package storetest

func matchErr(e1, e2 error) bool {
	return (e1 == nil && e2 == nil) || (e1 != nil && e2 != nil && e1.Error() == e2.Error())
}

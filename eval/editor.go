package eval

import "sync"

// Editor is the interface that the line editor has to satisfy. It is needed so
// that this package does not depend on the edit package.
type Editor interface {
	Active() bool
	ActiveMutex() *sync.Mutex
	Notify(string, ...interface{})
}

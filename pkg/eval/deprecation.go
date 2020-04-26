package eval

import (
	"sync"

	"github.com/elves/elvish/pkg/diag"
)

type deprecationRegistry struct {
	mutex      sync.Mutex
	registered map[deprecation]struct{}
}

func newDeprecationRegistry() deprecationRegistry {
	return deprecationRegistry{registered: make(map[deprecation]struct{})}
}

type deprecation struct {
	srcName  string
	location diag.Ranging
	message  string
}

// Registers a deprecation, and returns whether it was registered for the first
// time.
func (r *deprecationRegistry) register(dep deprecation) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.registered[dep]; ok {
		return false
	}
	r.registered[dep] = struct{}{}
	return true
}

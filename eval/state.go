package eval

import "sync"

// A bag of runtime states and configs exposed to Elvish code.
type state struct {
	mutex sync.RWMutex
	// The prefix to prepend to value outputs when writing them to terminal.
	valuePrefix string
	// Whether to notify the success of background jobs.
	notifyBgJobSuccess bool
	// The current number of background jobs.
	numBgJobs int
}

func (s *state) getValuePrefix() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.valuePrefix
}

func (s *state) getNotifyBgJobSuccess() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.notifyBgJobSuccess
}

func (s *state) getNumBgJobs() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.numBgJobs
}

func (s *state) addNumBgJobs(delta int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.numBgJobs += delta
}

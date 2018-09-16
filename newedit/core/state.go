package core

import "sync"

// State wraps RawState, providing methods for concurrency-safe access. The
// getter methods also paper over nil values to make the empty State value more
// usable.
type State struct {
	Raw   RawState
	Mutex sync.RWMutex
}

// CopyRaw returns a copy of the raw state.
func (s *State) CopyRaw() *RawState {
	raw := s.Raw
	return &raw
}

// Returns a finalized State, intended for use in the final redraw.
func (s *State) finalize() *RawState {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return &RawState{Mode: basicMode{}, Code: s.Raw.Code, Dot: len(s.Raw.Code)}
}

// Mode returns the current mode. If the internal mode value is nil, it returns
// a default Mode implementation.
func (s *State) Mode() Mode {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return getMode(s.Raw.Mode)
}

func getMode(m Mode) Mode {
	if m == nil {
		return basicMode{}
	}
	return m
}

// Code returns the code.
func (s *State) Code() string {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.Raw.Code
}

// CodeAndDot returns the code and dot of the state.
func (s *State) CodeAndDot() (string, int) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.Raw.Code, s.Raw.Dot
}

// CodeBeforeDot returns the part of code before the dot.
func (s *State) CodeBeforeDot() string {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.Raw.Code[:s.Raw.Dot]
}

// CodeBeforeDot returns the part of code after the dot.
func (s *State) CodeAfterDot() string {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.Raw.Code[s.Raw.Dot:]
}

// Resets resets the internal state to an empty value.
func (s *State) Reset() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Raw = RawState{}
}

type RawState struct {
	Mode    Mode
	Code    string
	Dot     int
	Pending *PendingCode
	Notes   []string
}

type PendingCode struct {
	Begin int
	End   int
	Text  string
}

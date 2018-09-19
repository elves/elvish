package prompt

import (
	"sync"
	"time"

	"github.com/elves/elvish/styled"
)

// Config wraps RawConfig for safe concurrent access. All of its methods are
// concurrency-safe, but direct field access must still be synchronized.
type Config struct {
	Raw   RawConfig
	Mutex sync.RWMutex
}

// Compute returns c.Raw.Compute while r-locking the mutex.
func (c *Config) Compute() func() styled.Text {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.Compute
}

// StaleTransform returns c.Raw.StaleTransform while r-locking the mutex.
func (c *Config) StaleTransform() func(styled.Text) styled.Text {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.StaleTransform
}

// StaleTransform returns c.Raw.StaleThreshold while r-locking the mutex.
func (c *Config) StaleThreshold() time.Duration {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.StaleThreshold
}

// Eagerness returns c.Raw.StaleThreshold while r-locking the mutex.
func (c *Config) Eagerness() int {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.Eagerness
}

// RawConfig keeps configurations for the prompt.
type RawConfig struct {
	// The function that computes the prompt.
	Compute func() styled.Text
	// Function to transform stale prompts.
	StaleTransform func(styled.Text) styled.Text
	// Threshold for a prompt to be considered as stale.
	StaleThreshold time.Duration
	// How eager the prompt should be updated. When >= 5, updated when directory
	// is changed. When >= 10, always update. Default is 5.
	Eagerness int
}

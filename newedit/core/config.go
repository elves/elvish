package core

import (
	"sync"
)

// Config wraps RawConfig for safe concurrent access. All of its methods are
// concurrency-safe, but direct field access must still be synchronized.
type Config struct {
	Raw   RawConfig
	Mutex sync.RWMutex
}

// MaxHeight returns c.Raw.MaxHeight while r-locking c.Mutex.
func (c *Config) MaxHeight() int {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.MaxHeight
}

// RawConfig keeps configurations of the editor.
type RawConfig struct {
	// Maximum lines of the terminal the editor may use. If MaxHeight <= 0,
	// there is no limit.
	MaxHeight int
	// Whether the rprompt is shown in the final redraw; in other words, whether
	// the rprompt persists in the terminal history when ReadCode returns.
	RPromptPersistent bool
}

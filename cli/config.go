package cli

import "sync"

type coreConfig struct {
	mu sync.Mutex

	maxHeight         int
	rpromptPersistent bool
}

func (cc *coreConfig) MaxHeight() int {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.maxHeight
}

func (cc *coreConfig) RPromptPersistent() bool {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.rpromptPersistent
}

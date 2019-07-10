package cli

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/histutil"
)

// Adaps App to implement the clicore.Config interface.
type coreConfig struct {
	*App
}

func (cc coreConfig) MaxHeight() (ret int) {
	cc.cfg.Mutex.Lock()
	defer cc.cfg.Mutex.Unlock()
	return cc.cfg.MaxHeight
}

func (cc coreConfig) RPromptPersistent() bool {
	cc.cfg.Mutex.Lock()
	defer cc.cfg.Mutex.Unlock()
	return cc.cfg.RPromptPersistent
}

func (cc coreConfig) BeforeReadline() {
	cc.cfg.Mutex.Lock()
	defer cc.cfg.Mutex.Unlock()
	for _, f := range cc.cfg.BeforeReadline {
		f()
	}
}

func (cc coreConfig) AfterReadline(code string) {
	cc.cfg.Mutex.Lock()
	defer cc.cfg.Mutex.Unlock()
	for _, f := range cc.cfg.AfterReadline {
		f(code)
	}
	if cc.cfg.HistoryStore != nil {
		_, err := cc.cfg.HistoryStore.AddCmd(histutil.Entry{Text: code})
		if err != nil {
			cc.core.Notify("db error: " + err.Error())
		}
	}
}

func (cc coreConfig) Highlighter() Highlighter {
	cc.cfg.Mutex.Lock()
	defer cc.cfg.Mutex.Unlock()
	return cc.cfg.Highlighter
}

func (cc coreConfig) Prompt() Prompt {
	cc.cfg.Mutex.Lock()
	defer cc.cfg.Mutex.Unlock()
	return cc.cfg.Prompt
}

func (cc coreConfig) RPrompt() Prompt {
	cc.cfg.Mutex.Lock()
	defer cc.cfg.Mutex.Unlock()
	return cc.cfg.RPrompt
}

func (cc coreConfig) InitMode() clitypes.Mode {
	cc.cfg.Mutex.Lock()
	defer cc.cfg.Mutex.Unlock()
	return cc.Insert
}

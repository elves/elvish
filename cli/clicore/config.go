package clicore

// Config is the configuration for an App.
type Config struct {
	MaxHeight         func() int
	RPromptPersistent func() bool
	BeforeReadline    func()
	AfterReadline     func(string)
	Highlighter       Highlighter
	Prompt            Prompt
	RPrompt           Prompt
}

func (cfg *Config) maxHeight() int {
	if cfg.MaxHeight != nil {
		return cfg.MaxHeight()
	}
	return -1
}

func (cfg *Config) rpromptPersistent() bool {
	if cfg.RPromptPersistent != nil {
		return cfg.RPromptPersistent()
	}
	return false
}

func (cfg *Config) beforeReadline() {
	if cfg.BeforeReadline != nil {
		cfg.BeforeReadline()
	}
}

func (cfg *Config) afterReadline(content string) {
	if cfg.AfterReadline != nil {
		cfg.AfterReadline(content)
	}
}

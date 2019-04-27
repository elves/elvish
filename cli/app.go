package cli

import (
	"github.com/elves/elvish/cli/clicore"
)

// App represents a CLI app.
type App = clicore.App

// AppConfig is a struct containing configurations for initializing an App.
type AppConfig struct {
	MaxHeight int

	BeforeReadline []func()
	AfterReadline  []func(string)

	Highlighter       Highlighter
	Prompt, RPrompt   Prompt
	RPromptPersistent bool

	InsertConfig InsertModeConfig
}

// NewApp creates a new App.
func NewApp(cfg *AppConfig) *App {
	app := clicore.NewAppFromStdIO()
	app.Config.Raw = clicore.RawConfig{
		MaxHeight:         cfg.MaxHeight,
		RPromptPersistent: cfg.RPromptPersistent,
	}
	app.BeforeReadline = cfg.BeforeReadline
	app.AfterReadline = cfg.AfterReadline
	app.Highlighter = cfg.Highlighter
	app.Prompt = cfg.Prompt
	app.RPrompt = cfg.RPrompt

	insertMode := newInsertMode(&cfg.InsertConfig, app.State())
	app.InitMode = insertMode
	return app
}

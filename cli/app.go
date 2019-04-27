package cli

import (
	"os"

	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/newedit/histlist"
	"github.com/elves/elvish/newedit/listing"
)

// App represents a CLI app.
type App struct {
	core     *clicore.App
	cfg      *AppConfig
	histlist *histlist.Mode
}

// AppConfig is a struct containing configurations for initializing an App.
type AppConfig struct {
	MaxHeight int

	BeforeReadline []func()
	AfterReadline  []func(string)

	Highlighter       Highlighter
	Prompt, RPrompt   Prompt
	RPromptPersistent bool

	HistoryStore HistoryStore

	InsertConfig InsertModeConfig
}

// NewAppFromStdIO creates a new App that reads from stdin and writes to stderr.
func NewAppFromStdIO(cfg *AppConfig) *App {
	return NewAppFromFiles(cfg, os.Stdin, os.Stderr)
}

// NewAppFromFiles creates a new App from the input and output files.
func NewAppFromFiles(cfg *AppConfig, in, out *os.File) *App {
	return NewApp(cfg, clicore.NewTTY(in, out), clicore.NewSignalSource())
}

// NewApp creates a new App.
func NewApp(cfg *AppConfig, t clicore.TTY, sigs clicore.SignalSource) *App {
	coreApp := clicore.NewApp(t, sigs)
	app := &App{
		core: coreApp,
		cfg:  cfg,
	}
	coreApp.Config.Raw = clicore.RawConfig{
		MaxHeight:         cfg.MaxHeight,
		RPromptPersistent: cfg.RPromptPersistent,
	}
	coreApp.BeforeReadline = cfg.BeforeReadline
	recordCmd := func(code string) {
		if cfg.HistoryStore == nil {
			return
		}
		err := cfg.HistoryStore.AddCmd(code)
		if err != nil {
			coreApp.Notify("db error: " + err.Error())
		}
	}
	afterReadline := append([]func(string){recordCmd}, cfg.AfterReadline...)
	coreApp.AfterReadline = afterReadline
	coreApp.Highlighter = cfg.Highlighter
	coreApp.Prompt = cfg.Prompt
	coreApp.RPrompt = cfg.RPrompt

	insertMode := newInsertMode(&cfg.InsertConfig, app)
	coreApp.InitMode = insertMode

	lsMode := &listing.Mode{}
	app.histlist = &histlist.Mode{Mode: lsMode}

	return app
}

// ReadCode causes the App to read from terminal.
func (app *App) ReadCode() (string, error) {
	return app.core.ReadCode()
}

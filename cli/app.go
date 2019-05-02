package cli

import (
	"os"

	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/cli/histlist"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/lastcmd"
	"github.com/elves/elvish/cli/listing"
)

// App represents a CLI app.
type App struct {
	core *clicore.App
	cfg  *AppConfig

	histlist *histlist.Mode
	lastcmd  *lastcmd.Mode
}

// AppConfig is a struct containing configurations for initializing an App.
type AppConfig struct {
	MaxHeight int

	BeforeReadline []func()
	AfterReadline  []func(string)

	Highlighter Highlighter

	Prompt, RPrompt   Prompt
	RPromptPersistent bool

	HistoryStore histutil.Store

	Wordifier Wordifier

	InsertModeConfig   InsertModeConfig
	HistlistModeConfig HistlistModeConfig
	LastcmdModeConfig  LastcmdModeConfig
}

// Wordifier is the type of a function that turns code into words.
type Wordifier func(code string) []string

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
		_, err := cfg.HistoryStore.AddCmd(histutil.Entry{Text: code})
		if err != nil {
			coreApp.Notify("db error: " + err.Error())
		}
	}
	afterReadline := append([]func(string){recordCmd}, cfg.AfterReadline...)
	coreApp.AfterReadline = afterReadline
	coreApp.Highlighter = cfg.Highlighter
	coreApp.Prompt = cfg.Prompt
	coreApp.RPrompt = cfg.RPrompt

	insertMode := newInsertMode(&cfg.InsertModeConfig, app)
	coreApp.InitMode = insertMode

	lsMode := &listing.Mode{}
	app.histlist = &histlist.Mode{
		Mode:       lsMode,
		KeyHandler: adaptBinding(cfg.HistlistModeConfig.Binding, app),
	}
	app.lastcmd = &lastcmd.Mode{
		Mode:       lsMode,
		KeyHandler: adaptBinding(cfg.LastcmdModeConfig.Binding, app),
	}

	return app
}

// ReadCode causes the App to read from terminal.
func (app *App) ReadCode() (string, error) {
	return app.core.ReadCode()
}

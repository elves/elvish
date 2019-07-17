package cli

import (
	"os"
	"sync"

	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/histlist"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/lastcmd"
	"github.com/elves/elvish/cli/listing"
)

// App represents a CLI app.
type App struct {
	core *clicore.App
	cfg  *AppConfig

	Insert   clitypes.Mode
	Listing  *listing.Mode
	Histlist *histlist.Mode
	Lastcmd  *lastcmd.Mode
}

// AppConfig is a struct containing configurations for initializing an App. It
// must not be copied once used.
type AppConfig struct {
	Mutex sync.RWMutex

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
	coreApp.Config = coreConfig{app}

	app.Insert = newInsertMode(app)
	lsMode := &listing.Mode{}
	app.Listing = lsMode
	app.Histlist = &histlist.Mode{
		Mode:       lsMode,
		KeyHandler: adaptBinding(cfg.HistlistModeConfig.Binding, app),
	}
	app.Lastcmd = &lastcmd.Mode{
		Mode:       lsMode,
		KeyHandler: adaptBinding(cfg.LastcmdModeConfig.Binding, app),
	}

	return app
}

// ReadCode causes the App to read from terminal.
func (app *App) ReadCode() (string, error) {
	return app.core.ReadCode()
}

// ReadCodeAsync is like ReadCode, but returns immediately with two channels
// that will get the return values of ReadCode. Useful in tests.
func (app *App) ReadCodeAsync() (<-chan string, <-chan error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := app.ReadCode()
		codeCh <- code
		errCh <- err
	}()
	return codeCh, errCh
}

// Notify adds a note and requests a redraw.
func (app *App) Notify(note string) {
	app.core.Notify(note)
}

// State returns the state of the App.
func (app *App) State() *clitypes.State {
	return app.core.State()
}

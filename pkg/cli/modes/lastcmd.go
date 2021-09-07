package modes

import (
	"fmt"
	"strconv"
	"strings"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

// Lastcmd is a mode for inspecting the last command, and inserting part of all
// of it. It is based on the ComboBox widget.
type Lastcmd interface {
	tk.ComboBox
}

// LastcmdSpec specifies the configuration for the lastcmd mode.
type LastcmdSpec struct {
	// Key bindings.
	Bindings tk.Bindings
	// Store provides the source for the last command.
	Store LastcmdStore
	// Wordifier breaks a command into words.
	Wordifier func(string) []string
}

// LastcmdStore is a subset of histutil.Store used in lastcmd mode.
type LastcmdStore interface {
	Cursor(prefix string) histutil.Cursor
}

var _ = LastcmdStore(histutil.Store(nil))

// NewLastcmd creates a new lastcmd mode.
func NewLastcmd(app cli.App, cfg LastcmdSpec) (Lastcmd, error) {
	codeArea, err := FocusedCodeArea(app)
	if err != nil {
		return nil, err
	}
	if cfg.Store == nil {
		return nil, errNoHistoryStore
	}
	c := cfg.Store.Cursor("")
	c.Prev()
	cmd, err := c.Get()
	if err != nil {
		return nil, fmt.Errorf("db error: %v", err)
	}
	wordifier := cfg.Wordifier
	if wordifier == nil {
		wordifier = strings.Fields
	}
	cmdText := cmd.Text
	words := wordifier(cmdText)
	entries := make([]lastcmdEntry, len(words)+1)
	entries[0] = lastcmdEntry{content: cmdText}
	for i, word := range words {
		entries[i+1] = lastcmdEntry{strconv.Itoa(i), strconv.Itoa(i - len(words)), word}
	}

	accept := func(text string) {
		codeArea.MutateState(func(s *tk.CodeAreaState) {
			s.Buffer.InsertAtDot(text)
		})
		app.PopAddon()
	}
	w := tk.NewComboBox(tk.ComboBoxSpec{
		CodeArea: tk.CodeAreaSpec{Prompt: modePrompt(" LASTCMD ", true)},
		ListBox: tk.ListBoxSpec{
			Bindings: cfg.Bindings,
			OnAccept: func(it tk.Items, i int) {
				accept(it.(lastcmdItems).entries[i].content)
			},
		},
		OnFilter: func(w tk.ComboBox, p string) {
			items := filterLastcmdItems(entries, p)
			if len(items.entries) == 1 {
				accept(items.entries[0].content)
			} else {
				w.ListBox().Reset(items, 0)
			}
		},
	})
	return w, nil
}

type lastcmdItems struct {
	negFilter bool
	entries   []lastcmdEntry
}

type lastcmdEntry struct {
	posIndex string
	negIndex string
	content  string
}

func filterLastcmdItems(allEntries []lastcmdEntry, p string) lastcmdItems {
	if p == "" {
		return lastcmdItems{false, allEntries}
	}
	var entries []lastcmdEntry
	negFilter := strings.HasPrefix(p, "-")
	for _, entry := range allEntries {
		if (negFilter && strings.HasPrefix(entry.negIndex, p)) ||
			(!negFilter && strings.HasPrefix(entry.posIndex, p)) {
			entries = append(entries, entry)
		}
	}
	return lastcmdItems{negFilter, entries}
}

func (it lastcmdItems) Show(i int) ui.Text {
	index := ""
	entry := it.entries[i]
	if it.negFilter {
		index = entry.negIndex
	} else {
		index = entry.posIndex
	}
	// NOTE: We now use a hardcoded width of 3 for the index, which will work as
	// long as the command has less than 1000 words (when filter is positive) or
	// 100 words (when filter is negative).
	return ui.T(fmt.Sprintf("%3s %s", index, entry.content))
}

func (it lastcmdItems) Len() int { return len(it.entries) }

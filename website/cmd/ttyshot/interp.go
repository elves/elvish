package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/creack/pty"
	"src.elv.sh/pkg/ui"
)

const (
	terminalRows = 100
	terminalCols = 60
)

var promptMarker = "[PROMPT]"

//go:embed rc.elv
var rcElv string

// Create a hermetic environment for generating a ttyshot. We want to ensure we don't use the real
// home directory, or interactive history, of the person running this tool.
func setupHome() (string, error) {
	homePath, err := os.MkdirTemp("", "ttyshot-*")
	if err != nil {
		return "", fmt.Errorf("create temp home: %w", err)
	}

	// The temporary directory may include symlinks in the path. Expand them so
	// that commands like tilde-abbr behaves as expected.
	resolvedHomePath, err := filepath.EvalSymlinks(homePath)
	if err != nil {
		return homePath, fmt.Errorf("resolve symlinks in homePath: %w", err)
	}
	homePath = resolvedHomePath

	err = ApplyDir(Dir{
		// Directories to be used in navigation mode.
		"bash": Dir{},
		"elvish": Dir{
			"0.x.0-release-notes.md": "This is the draft release notes for 0.x.0.",
			"CONTRIBUTING.md":        "",
			"Dockerfile":             "",
			"LICENSE":                "",
			"Makefile":               "",
			"PACKAGING.md":           "",
			"README.md":              "",
			"SECURITY.md":            "",
			"cmd":                    Dir{},
			"go.mod":                 "",
			"go.sum":                 "",
			"pkg":                    Dir{},
			"syntaxes":               Dir{},
			"tools":                  Dir{},
			"vscode":                 Dir{},
			"website":                Dir{},
		},
		"zsh": Dir{},

		// Will keep tmux and elvish's sockets, and raw output of capture-pane
		".tmp": Dir{},

		".config": Dir{
			"elvish": Dir{
				"rc.elv": rcElv,
			},
		},
	}, homePath)
	return homePath, err
}

func createTtyshot(homePath string, script []demoOp, saveRaw string) ([]byte, error) {
	ctrl, tty, err := pty.Open()
	if err != nil {
		return nil, err
	}
	defer ctrl.Close()
	defer tty.Close()
	winsize := pty.Winsize{Rows: terminalRows, Cols: terminalCols}
	pty.Setsize(ctrl, &winsize)

	// Relay the output of the ttyshot Elvish session to the channel that will capture and evaluate
	// the output; e.g., to detect whether a prompt was seen.
	ttyOutput := make(chan byte, 32*1024)
	go func() {
		for {
			content := make([]byte, 1024)
			n, err := ctrl.Read(content)
			if n == 0 {
				close(ttyOutput)
				return
			}
			if err != nil {
				panic(err)
			}
			for i := 0; i < n; i++ {
				ttyOutput <- content[i]
			}
		}
	}()

	rawPath := filepath.Join(homePath, ".tmp", "ttyshot.raw")
	if saveRaw != "" {
		saveRaw, err := filepath.Abs(saveRaw)
		if err != nil {
			return nil, fmt.Errorf("resolve path to raw dump file: %w", err)
		}
		os.Symlink(saveRaw, rawPath)
	}

	doneCh, err := spawnElvish(homePath, tty)
	if err != nil {
		return nil, err
	}
	executeScript(script, ctrl, ttyOutput)

	err = <-doneCh
	if err != nil {
		return nil, err
	}

	rawBytes, err := os.ReadFile(rawPath)
	if err != nil {
		return nil, err
	}

	ttyshot := string(rawBytes)
	ttyshot = strings.TrimRight(ttyshot, "\n")
	ttyshot = strings.ReplaceAll(ttyshot, promptMarker+"\n", "")
	return []byte(sgrTextToHTML(ttyshot) + "\n"), nil
}

func spawnElvish(homePath string, tty *os.File) (<-chan error, error) {
	// Construct a file name for the tmux and Elvish daemon socket files in the temp home path.
	tmuxSock := filepath.Join(homePath, ".tmp", "tmux.sock")
	elvSock := filepath.Join(homePath, ".tmp", "elv.sock")

	elvishPath, err := exec.LookPath("elvish")
	if err != nil {
		return nil, fmt.Errorf("find elvish: %w", err)
	}
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return nil, fmt.Errorf("find tmux: %w", err)
	}

	// Start tmux and have it start the hermetic Elvish shell.
	tmuxCmd := exec.Cmd{
		Path: tmuxPath,
		Args: []string{
			tmuxPath,
			"-S", tmuxSock, "-f", "/dev/null", "-u", "-T", "256,RGB",
			"new-session", elvishPath, "-sock", elvSock},
		Dir: homePath,
		Env: []string{
			"HOME=" + homePath,
			"PATH=" + os.Getenv("PATH"),
			// The actual value doesn't matter here, as long as it can be looked
			// up in terminfo. We rely on the -T flag above to force tmux to
			// support certain terminal features.
			"TERM=xterm",
		},
		Stdin:  tty,
		Stdout: tty,
		Stderr: tty,
	}

	doneCh := make(chan error)
	go func() {
		doneCh <- tmuxCmd.Run()
	}()

	return doneCh, nil
}

func executeScript(script []demoOp, ctrl *os.File, ttyOutput chan byte) {
	implicitEnter := true
	nextCmdNum := 1
	for _, op := range script {
		switch op.what {
		case opText:
			text := op.val.([]byte)
			ctrl.Write(text)
			if implicitEnter {
				ctrl.Write([]byte{'\r'})
			}
		case opAlt:
			ctrl.Write([]byte{'\033', op.val.(byte)})
		case opCtrl:
			ctrl.Write([]byte{op.val.(byte) & 0x1F})
		case opEnter:
			ctrl.Write([]byte{'\r'})
			implicitEnter = true
		case opUp:
			ctrl.Write([]byte{'\033', '[', 'A'})
		case opDown:
			ctrl.Write([]byte{'\033', '[', 'B'})
		case opRight:
			ctrl.Write([]byte{'\033', '[', 'C'})
		case opLeft:
			ctrl.Write([]byte{'\033', '[', 'D'})
		case opNoEnter:
			implicitEnter = false
		case opSleep:
			time.Sleep(op.val.(time.Duration))
		case opWaitForPrompt:
			waitForOutput(ttyOutput, promptMarker,
				func(content []byte) bool { return bytes.Contains(content, []byte(promptMarker)) })
			nextCmdNum++
		case opWaitForString:
			expected := op.val.([]byte)
			waitForOutput(ttyOutput, string(expected),
				func(content []byte) bool { return bytes.Contains(content, expected) })
		case opWaitForRegexp:
			expected := op.val.(*regexp.Regexp)
			waitForOutput(ttyOutput, expected.String(),
				func(content []byte) bool { return expected.Match(content) })
		default:
			panic("unhandled op")
		}
	}
	// Alt-q is bound to a function that captures the content of the pane and
	// exits
	ctrl.Write([]byte{'\033', 'q'})
}

func waitForOutput(ttyOutput chan byte, expected string, matcher func([]byte) bool) []byte {
	text := make([]byte, 0, 4096)
	// It shouldn't take more than a couple of seconds to see the expected output so use a timeout
	// an order of magnitude longer to allow for overloaded systems.
	timeout := time.After(30 * time.Second)
	for {
		var newByte byte
		select {
		case newByte = <-ttyOutput:
		case <-timeout:
			fmt.Fprintf(os.Stderr, "Timeout waiting for text matching: %q\n", expected)
			fmt.Fprintf(os.Stderr, "This is what we've captured so far:\n%q\n", text)
			os.Exit(3)
		}
		text = append(text, newByte)
		if matcher(text) {
			break
		}
	}
	return text
}

func sgrTextToHTML(ttyshot string) string {
	t := ui.ParseSGREscapedText(ttyshot)

	var sb strings.Builder
	for _, c := range t {
		var classes []string
		for _, c := range c.Style.SGRValues() {
			classes = append(classes, "sgr-"+c)
		}
		text, newline := c.Text, false
		if c.Text[len(c.Text)-1] == '\n' {
			newline = true
			text = c.Text[:len(c.Text)-1]
		}
		fmt.Fprintf(&sb,
			`<span class="%s">%s</span>`, strings.Join(classes, " "), html.EscapeString(text))
		if newline {
			sb.Write([]byte{'\n'})
		}
	}

	return sb.String()
}

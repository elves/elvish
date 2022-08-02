//go:build !windows && !plan9 && !js

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/creack/pty"
	"src.elv.sh/pkg/sys/eunix"
	"src.elv.sh/pkg/ui"
)

const (
	terminalRows = 100
	terminalCols = 60
)

var promptMarker = "[PROMPT]"

//go:embed rc.elv
var rcElv string

// Creates a temporary home directory for running tmux and elvish in. The caller
// is responsible for removing the directory.
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

func createTtyshot(homePath string, script []op, saveRaw string) ([]byte, error) {
	ctrl, tty, err := pty.Open()
	if err != nil {
		return nil, err
	}
	defer ctrl.Close()
	defer tty.Close()
	winsize := pty.Winsize{Rows: terminalRows, Cols: terminalCols}
	pty.Setsize(ctrl, &winsize)

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
	executeScript(script, ctrl)

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
	elvishPath, err := exec.LookPath("elvish")
	if err != nil {
		return nil, fmt.Errorf("find elvish: %w", err)
	}
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return nil, fmt.Errorf("find tmux: %w", err)
	}

	tmuxSock := filepath.Join(homePath, ".tmp", "tmux.sock")
	elvSock := filepath.Join(homePath, ".tmp", "elv.sock")

	// Start tmux and have it start a hermetic Elvish session.
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

func executeScript(script []op, ctrl *os.File) {
	implicitEnter := true
	nextCmdNum := 1
	for _, op := range script {
		switch op.typ {
		case opText:
			text := op.val.(string)
			ctrl.WriteString(text)
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
		case opWaitForPrompt:
			waitForOutput(ctrl, promptMarker,
				func(content []byte) bool { return bytes.Contains(content, []byte(promptMarker)) })
			nextCmdNum++
		case opWaitForString:
			expected := op.val.(string)
			waitForOutput(ctrl, expected,
				func(content []byte) bool { return bytes.Contains(content, []byte(expected)) })
		case opWaitForRegexp:
			expected := op.val.(*regexp.Regexp)
			waitForOutput(ctrl, expected.String(),
				func(content []byte) bool { return expected.Match(content) })
		default:
			panic("unhandled op")
		}
	}
	// Alt-q is bound to a function that captures the content of the pane and
	// exits
	ctrl.Write([]byte{'\033', 'q'})
}

func waitForOutput(f *os.File, expected string, matcher func([]byte) bool) error {
	var buf bytes.Buffer
	// It shouldn't take more than a couple of seconds to see the expected
	// output, so use a timeout an order of magnitude longer to allow for
	// overloaded systems.
	deadline := time.Now().Add(30 * time.Second)
	for {
		budget := time.Until(deadline)
		if budget <= 0 {
			break
		}
		ready, err := eunix.WaitForRead(budget, f)
		if err != nil {
			return fmt.Errorf("waiting for tmux output: %w", err)
		}
		if !ready[0] {
			break
		}
		_, err = io.CopyN(&buf, f, 1)
		if err != nil {
			return fmt.Errorf("reading tmux output: %w", err)
		}
		if matcher(buf.Bytes()) {
			return nil
		}
	}
	return fmt.Errorf("timed out waiting for %s in tmux output; output so far: %q", expected, buf)
}

var htmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

func sgrTextToHTML(ttyshot string) string {
	t := ui.ParseSGREscapedText(ttyshot)

	var sb strings.Builder
	for i, line := range t.SplitByRune('\n') {
		if i > 0 {
			sb.WriteRune('\n')
		}
		for j, seg := range line {
			var classes []string
			for _, c := range seg.Style.SGRValues() {
				classes = append(classes, "sgr-"+c)
			}
			text := seg.Text
			// We pass -N to tmux capture-pane in order to correctly preserve
			// trailing spaces that have background colors. However, this
			// preserves unstyled trailing spaces too, which makes the ttyshot
			// harder to copy-paste, so strip it.
			if len(classes) == 0 && j == len(line)-1 {
				text = strings.TrimRight(text, " ")
			}
			if text == "" {
				continue
			}
			escapedText := htmlEscaper.Replace(text)
			if len(classes) == 0 {
				sb.WriteString(escapedText)
			} else {
				fmt.Fprintf(&sb, `<span class="%s">%s</span>`, strings.Join(classes, " "), escapedText)
			}
		}
	}

	return sb.String()
}

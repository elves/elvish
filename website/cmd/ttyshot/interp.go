//go:build unix

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"src.elv.sh/pkg/sys/eunix"
	"src.elv.sh/pkg/ui"
)

const (
	cutMarker    = "[CUT]"
	promptMarker = "[PROMPT]"
)

// "tmux capture-pane" can save superfluous trailing spaces, so when removing
// these patterns we need to account for that.
var (
	cutPattern    = regexp.MustCompile(regexp.QuoteMeta(cutMarker) + " *\n")
	promptPattern = regexp.MustCompile(regexp.QuoteMeta(promptMarker) + " *\n")
)

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
			"1.0-release.md":  "1.0 has not been released yet.",
			"CONTRIBUTING.md": "",
			"Dockerfile":      "",
			"LICENSE":         "",
			"Makefile":        "",
			"PACKAGING.md":    "",
			"README.md":       "",
			"SECURITY.md":     "",
			"cmd":             Dir{},
			"go.mod":          "",
			"go.sum":          "",
			"pkg":             Dir{},
			"syntaxes":        Dir{},
			"tools":           Dir{},
			"vscode":          Dir{},
			"website":         Dir{},
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

func createTtyshot(homePath string, script *script, saveRaw string) ([]byte, error) {
	ctrl, tty, err := pty.Open()
	if err != nil {
		return nil, err
	}
	defer ctrl.Close()
	defer tty.Close()
	winsize := pty.Winsize{Rows: script.rows, Cols: script.cols}
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
	finalEmptyPrompt := executeScript(script.ops, ctrl, homePath)
	log.Println("executed script, waiting for tmux to exit")

	// Drain outputs from the terminal. This is needed so that tmux can exit
	// properly without blocking on flushing outputs.
	go io.Copy(io.Discard, ctrl)

	err = <-doneCh
	if err != nil {
		return nil, err
	}

	rawBytes, err := os.ReadFile(rawPath)
	if err != nil {
		return nil, err
	}

	ttyshot := string(rawBytes)
	// Remove all content before the last cutMarker.
	segments := cutPattern.Split(ttyshot, -1)
	ttyshot = segments[len(segments)-1]

	// Strip all the prompt markers and the final empty prompt.
	segments = promptPattern.Split(ttyshot, -1)
	if finalEmptyPrompt {
		segments = segments[:len(segments)-1]
	}
	ttyshot = strings.Join(segments, "")

	ttyshot = strings.TrimRight(ttyshot, "\n")
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
	log.Println("started tmux, socket", tmuxSock)

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- tmuxCmd.Run()
		log.Println("tmux exited")
	}()

	return doneCh, nil
}

func executeScript(script []op, ctrl *os.File, homePath string) (finalEmptyPrompt bool) {
	for _, op := range script {
		log.Println("waiting for prompt")
		err := waitForPrompt(ctrl)
		if err != nil {
			// TODO: Handle the error properly
			panic(err)
		}

		log.Println("executing", op)
		if op.isTmux {
			tmuxSock := filepath.Join(homePath, ".tmp", "tmux.sock")
			tmuxCmd := exec.Command("tmux",
				append([]string{"-S", tmuxSock}, strings.Fields(op.code)...)...)
			tmuxCmd.Env = []string{}
			err := tmuxCmd.Run()
			if err != nil {
				// TODO: Handle the error properly
				panic(err)
			}
		} else {
			for i, line := range strings.Split(op.code, "\n") {
				if i > 0 {
					// Use Alt-Enter to avoid committing the code
					ctrl.WriteString("\033\r")
				}
				ctrl.WriteString(line)
			}
			ctrl.WriteString("\r")
		}
	}

	if len(script) > 0 && !script[len(script)-1].isTmux {
		log.Println("waiting for final empty prompt")
		finalEmptyPrompt = true
		err := waitForPrompt(ctrl)
		if err != nil {
			// TODO: Handle the error properly
			panic(err)
		}
	}

	log.Println("sending Alt-q")
	// Alt-q is bound to a function that captures the content of the pane and
	// exits
	ctrl.Write([]byte{'\033', 'q'})
	return finalEmptyPrompt
}

func waitForPrompt(f *os.File) error {
	return waitForOutput(f, promptMarker,
		func(bs []byte) bool { return bytes.HasSuffix(bs, []byte(promptMarker)) })
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
			if err == syscall.EINTR {
				continue
			}
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

// We use this instead of html.EscapeString, since the latter also escapes ' and
// " unnecessarily.
var htmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

func sgrTextToHTML(ttyshot string) string {
	t := ui.ParseSGREscapedText(ttyshot)

	var sb strings.Builder
	for i, line := range t.SplitByRune('\n') {
		if i > 0 {
			sb.WriteRune('\n')
		}
		for j, seg := range line {
			style := seg.Style
			var classes []string
			if style.Inverse {
				// The inverse attribute means that the foreground and
				// background colors should be swapped, which cannot be
				// expressed in pure CSS. To work around this, this code swaps
				// the foreground and background colors, and uses two special
				// CSS classes to indicate that the foreground/background should
				// take the inverse of the default color.
				style.Inverse = false
				style.Fg, style.Bg = style.Bg, style.Fg
				if style.Fg == nil {
					classes = append(classes, "sgr-7fg")
				}
				if style.Bg == nil {
					classes = append(classes, "sgr-7bg")
				}
			}

			for _, c := range style.SGRValues() {
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

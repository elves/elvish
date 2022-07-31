package main

import (
	"bytes"
	"embed"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/creack/pty"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/ui"
)

const (
	terminalRows = 100
	terminalCols = 60
)

var promptRe = regexp.MustCompile(`^\[\d+\]$`)
var promptFmt = "[%d]"

//go:embed cp-elvish.sh tmux.conf rc.elv
var assets embed.FS

// Create a hermetic environment for generating a ttyshot. We want to ensure we don't use the real
// home directory, or interactive history, of the person running this tool.
func initEnv() (string, string, func(), error) {
	// There are systems, such as macOs, which generate a temp dir that includes symlinks in the
	// path. For example, `/var/` => `/private/var`. Expand those symlinks so that Elvish command
	// `tilde-abbr` will behave as expected.
	homePath, err := os.MkdirTemp("", "ttyshot-*")
	if err != nil {
		return "", "", nil, fmt.Errorf("create temp home: %w", err)
	}
	homePath, err = filepath.EvalSymlinks(homePath)
	if err != nil {
		return "", "", nil, fmt.Errorf("resolve symlinks in homePath: %w", err)
	}
	// We'll put the Elvish and Tmux socket files in this directory. This makes the "navigation"
	// mode ttyshots a trifle less confusing.
	tmp := filepath.Join(homePath, "tmp")
	os.Mkdir(tmp, 0o700)

	entries, _ := assets.ReadDir(".")
	for _, entry := range entries {
		name := entry.Name()
		content, _ := assets.ReadFile(name)
		err := os.WriteFile(filepath.Join(tmp, name), content, 0o700)
		if err != nil {
			return "", "", nil, fmt.Errorf("write embedded file %q: %w", name, err)
		}
	}

	// We don't pass any XDG env vars to the Elvish programs we spawn. We want them to rely solely
	// on HOME in order to force using our hermetic home.
	os.Setenv("HOME", homePath)
	os.Unsetenv(env.XDG_CONFIG_HOME)
	os.Unsetenv(env.XDG_DATA_DIRS)
	os.Unsetenv(env.XDG_DATA_HOME)
	os.Unsetenv(env.XDG_STATE_HOME)
	os.Unsetenv(env.XDG_RUNTIME_DIR)

	// Create the Elvish local state directory in the hermetic home.
	dotLocalStateElvish := filepath.Join(homePath, ".local", "state", "elvish")
	if err := os.MkdirAll(dotLocalStateElvish, 0o700); err != nil {
		return "", "", nil, fmt.Errorf("create state dir: %w", err)
	}

	// Copy the Elvish source code to the hermetic home for use in demos of things like Elvish's
	// "navigation" mode.
	copySrcPath := filepath.Join(tmp, "cp-elvish.sh")
	copySrcCmd := exec.Cmd{
		Path: copySrcPath,
		Args: []string{copySrcPath, homePath},
	}
	if err := copySrcCmd.Run(); err != nil {
		return "", "", nil, err
	}

	// Create a couple of other directories to make demos of "navigation" mode more interesting.
	os.Mkdir(filepath.Join(homePath, "bash"), 0o700)
	os.Mkdir(filepath.Join(homePath, "zsh"), 0o700)

	// Ensure the terminal type seen by tmux is a widely recognized terminal definition. This makes
	// it possible to generate ttyshots in a continuous deployment environment. It's also good to
	// decouple invocations from an environment we don't control if this is run by hand from an
	// interactive shell whose TERM value we can't predict.
	_ = os.Setenv("TERM", "xterm-256color")

	cleanup := func() {
		if err := os.RemoveAll(homePath); err != nil {
			fmt.Fprintln(os.Stderr, "Warning: unable to remove temp HOME:", err.Error())
		}
	}

	dbPath := filepath.Join(dotLocalStateElvish, "db.bolt")
	return homePath, dbPath, cleanup, nil
}

func createTtyshot(homePath, dbPath string, script []demoOp, outFile, rawSave *os.File) error {
	ctrl, tty, err := pty.Open()
	if err != nil {
		return err
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

	doneCh, err := spawnElvish(homePath, dbPath, tty)
	if err != nil {
		return err
	}
	_, err = executeScript(script, ctrl, ttyOutput)
	if err != nil {
		return err
	}

	err = <-doneCh
	if err != nil {
		return err
	}

	ttyshotBytes, err := os.ReadFile(filepath.Join(homePath, "tmp", "ttyshot.raw"))
	if err != nil {
		return err
	}
	ttyshot := string(ttyshotBytes)

	rawSave.Write(ttyshotBytes)
	// Trim the last, or all, trailing newlines in order to eliminate from the generated HTML
	// unwanted empty lines at the bottom of the ttyshot. The latter behavior occurs if the ttyshot
	// specification includes the `trim-empty` directive.
	ttyshot = strings.TrimRight(ttyshot, "\n")
	outFile.WriteString(sgrTextToHTML(ttyshot))
	outFile.WriteString("\n")
	return nil
}

func spawnElvish(homePath, dbPath string, tty *os.File) (<-chan error, error) {
	// Construct a file name for the tmux and Elvish daemon socket files in the temp home path.
	tmuxSock := filepath.Join(homePath, "tmp", "tmux.sock")
	elvSock := filepath.Join(homePath, "tmp", "elv.sock")

	elvishPath, err := exec.LookPath("elvish")
	if err != nil {
		return nil, fmt.Errorf("find elvish: %w", err)
	}
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return nil, fmt.Errorf("find tmux: %w", err)
	}

	// Start tmux and have it start the hermetic Elvish shell.
	elvRcPath := filepath.Join(homePath, "tmp", "rc.elv")
	tmuxCmd := exec.Cmd{
		Path: tmuxPath,
		Args: []string{
			tmuxPath,
			"-S", tmuxSock,
			"-f", filepath.Join(homePath, "tmp", "tmux.conf"),
			"new-session",
			"-s", "ttyshot",
			"-c", homePath,
			elvishPath, "-rc", elvRcPath, "-sock", elvSock},
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

func executeScript(script []demoOp, ctrl *os.File, ttyOutput chan byte) (bool, error) {
	trimEmptyLines := false
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
			expected := fmt.Sprintf(promptFmt, nextCmdNum)
			waitForOutput(ttyOutput, expected,
				func(content []byte) bool { return bytes.Contains(content, []byte(expected)) })
			nextCmdNum++
		case opWaitForString:
			expected := op.val.([]byte)
			waitForOutput(ttyOutput, string(expected),
				func(content []byte) bool { return bytes.Contains(content, expected) })
		case opWaitForRegexp:
			expected := op.val.(*regexp.Regexp)
			waitForOutput(ttyOutput, expected.String(),
				func(content []byte) bool { return expected.Match(content) })
		case opTrimEmptyLines:
			trimEmptyLines = true
		default:
			panic("unhandled op")
		}
	}
	// Alt-q is bound to a function that captures the content of the pane and
	// exits
	ctrl.Write([]byte{'\033', 'q'})
	return trimEmptyLines, nil
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
		// This "undoes" the ugly hack in rc.elv that requires we gratuitously
		// modify the style of the prompt to make it practical to recognize a prompt when executing
		// a ttyshot script.
		if promptRe.Match([]byte(text)) {
			// It looks like the text might be a shell prompt. Check if the styling matches the case
			// that needs to be fixed.
			if len(classes) >= 1 && classes[0] == "sgr-90" {
				classes[0] = "sgr-30" // fg-bright-black => fg-black
			}
		}

		fmt.Fprintf(&sb,
			`<span class="%s">%s</span>`, strings.Join(classes, " "), html.EscapeString(text))
		if newline {
			sb.Write([]byte{'\n'})
		}
	}

	return sb.String()
}

// Generate a ttyshot HTML image from a ttyshot specification.
//
// Usage: ./ttyshot website/ttyshot/*.spec
//
// You can recreate all the ttyshots by running the following from the project top-level directory:
//
//   make ttyshot
//   for f [website/ttyshot/**.spec] { put $f; ./ttyshot $f >/dev/tty 2>&1 }
//
// This assumes working `elvish` and `tmux` programs in $E:PATH.
//
package main

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/creack/pty"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/ui"
)

// Note: This depends on a custom tmux.rc that disables the status line. Otherwise the simulated tty
// would need to have 17 rows to achieve the desired snapshot dimensions.
const (
	terminalRows = 16
	terminalCols = 60
)

// Operations for driving a demo ttyshot.
const (
	opEnter          = iota // enable implicit Enter key and send an Enter key
	opNoEnter               // inhibit implicit Enter key
	opTrimEmptyLines        // trim trailing empty lines -- can occur anywhere in the spec
	opUp                    // send Up arrow sequence
	opDown                  // send Down arrow sequence
	opRight                 // send Right arrow sequence
	opLeft                  // send Left arrow sequence
	opText                  // send the provided text, optionally followed by Enter
	opAlt                   // send an alt sequence
	opCtrl                  // send a control character
	opSleep                 // sleep for the specified duration
	opWaitForPrompt         // wait for the expected "<n>" (command number) in the next prompt
	opWaitForRegexp         // wait for sequence of bytes matching the regexp
	opWaitForString         // wait for the literal sequence of bytes
)

type demoOp struct {
	what int
	val  any
}

var promptRe = regexp.MustCompile(`^\[\d+\]$`)
var promptFmt = "[%d]"

func parseDirective(directive []byte) (demoOp, error) {
	if bytes.HasPrefix(directive, []byte("sleep ")) {
		duration, err := time.ParseDuration(string(directive[6:]) + "s")
		if err != nil {
			return demoOp{}, err
		}
		return demoOp{opSleep, duration}, nil
	}

	if bytes.Equal(directive, []byte("no-enter")) {
		return demoOp{opNoEnter, nil}, nil
	}

	if bytes.Equal(directive, []byte("trim-empty")) {
		return demoOp{opTrimEmptyLines, nil}, nil
	}

	if bytes.Equal(directive, []byte("enter")) {
		return demoOp{opEnter, nil}, nil
	}

	// Tab is frequently used so it's useful to support it as a directive rather than requiring
	// `//ctrl I`.
	if bytes.Equal(directive, []byte("tab")) {
		return demoOp{opCtrl, byte('I')}, nil
	}

	if bytes.HasPrefix(directive, []byte("ctrl ")) {
		if len(directive) != 6 {
			return demoOp{}, errors.New("invalid ctrl directive: " + string(directive))
		}
		return demoOp{opCtrl, directive[5]}, nil
	}

	if bytes.HasPrefix(directive, []byte("alt ")) {
		if len(directive) != 5 {
			return demoOp{}, errors.New("invalid alt directive: " + string(directive))
		}
		return demoOp{opAlt, directive[4]}, nil
	}

	if bytes.Equal(directive, []byte("prompt")) {
		return demoOp{opWaitForPrompt, nil}, nil
	}

	if bytes.Equal(directive, []byte("up")) {
		return demoOp{opUp, nil}, nil
	}

	if bytes.Equal(directive, []byte("down")) {
		return demoOp{opDown, nil}, nil
	}

	if bytes.Equal(directive, []byte("right")) {
		return demoOp{opRight, nil}, nil
	}

	if bytes.Equal(directive, []byte("left")) {
		return demoOp{opLeft, nil}, nil
	}

	if bytes.HasPrefix(directive, []byte("wait-for-re ")) {
		re, err := regexp.Compile(string(directive[12:]))
		if err != nil {
			return demoOp{}, errors.New("invalid wait-for-re value: " + string(directive[12:]))
		}
		return demoOp{opWaitForRegexp, re}, nil
	}

	if bytes.HasPrefix(directive, []byte("wait-for-str ")) {
		return demoOp{opWaitForString, directive[13:]}, nil
	}

	return demoOp{}, errors.New("unrecognized directive: " + string(directive))
}

func parseSpec(content []byte) ([]demoOp, error) {
	lines := bytes.Split(content, []byte{'\n'})
	ops := make([]demoOp, 1, len(lines)+2)
	ops[0] = demoOp{opWaitForPrompt, nil}

	for _, line := range lines {
		if len(line) == 0 {
			continue // ignore empty lines
		}
		if bytes.HasPrefix(line, []byte("//")) {
			directive, err := parseDirective(line[2:])
			if err != nil {
				return ops, err
			}
			ops = append(ops, directive)
		} else {
			ops = append(ops, demoOp{opText, line})
		}
	}

	return ops, nil
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
		// This "undoes" the ugly hack in website/ttyshot/ttyshot.rc that requires we gratuitously
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

var cmdNum int = 0

func executeScript(script []demoOp, master *os.File, ttyOutput chan byte) (bool, error) {
	trimEmptyLines := false
	implicitEnter := true
	for _, op := range script {
		switch op.what {
		case opText:
			text := op.val.([]byte)
			master.Write(text)
			if implicitEnter {
				master.Write([]byte{'\r'})
			}
		case opAlt:
			master.Write([]byte{'\033', op.val.(byte)})
		case opCtrl:
			master.Write([]byte{op.val.(byte) & 0x1F})
		case opEnter:
			master.Write([]byte{'\r'})
			implicitEnter = true
		case opUp:
			master.Write([]byte{'\033', '[', 'A'})
		case opDown:
			master.Write([]byte{'\033', '[', 'B'})
		case opRight:
			master.Write([]byte{'\033', '[', 'C'})
		case opLeft:
			master.Write([]byte{'\033', '[', 'D'})
		case opNoEnter:
			implicitEnter = false
		case opSleep:
			time.Sleep(op.val.(time.Duration))
		case opWaitForPrompt:
			cmdNum++
			expected := fmt.Sprintf(promptFmt, cmdNum)
			waitForOutput(ttyOutput, expected,
				func(content []byte) bool { return bytes.Contains(content, []byte(expected)) })
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
	return trimEmptyLines, nil
}

func spawnElvish(homePath, dbPath string, slave *os.File,
	ttyImage *bytes.Buffer) (chan bool, chan bool) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("unable to determine the CWD: " + err.Error())
	}

	triggerTtyCapture := make(chan bool)
	ttyCaptureDone := make(chan bool)

	// Construct a file name for the tmux and Elvish daemon socket files in the temp home path.
	tmuxSock := filepath.Join(homePath, "tmp", "tmux.sock")
	elvSock := filepath.Join(homePath, "tmp", "elv.sock")

	elvishPath, err := exec.LookPath("elvish")
	if err != nil {
		log.Fatal("unable to find elvish: " + err.Error())
	}
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		log.Fatal("unable to find tmux: " + err.Error())
	}

	devnul, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		log.Fatal("unable to open os.DevNull: " + err.Error())
	}

	daemonCmd := exec.Cmd{
		Path:   elvishPath,
		Args:   []string{"elvish", "-daemon", "-sock", elvSock, "-db", dbPath},
		Stdin:  devnul,
		Stdout: devnul,
		Stderr: devnul,
	}
	// Run the Elvish daemon using the hermetic home.
	go func() {
		if err := daemonCmd.Start(); err != nil {
			log.Fatal(err)
		}
		// The daemon will exit when the Elvish shell we're intereacting with exits. So we simply
		// need to wait for the daemon to terminate.
		if err := daemonCmd.Wait(); err != nil {
			log.Fatal(err)
		}
	}()
	// Wait for the Elvish daemon to create the socket file before we start the Elvish shell.
	// This isn't strictly speaking necessary but helps avoid race conditions and makes it more
	// likely we'll abort with a useful diagnostic if the daemon fails to start.
	launchTime := time.Now()
	time.Sleep(10 * time.Millisecond)
	for {
		if _, err := os.Stat(elvSock); err == nil {
			break
		}
		if time.Now().Sub(launchTime) > time.Duration(5*time.Second) {
			log.Fatal("Elvish daemon failed to create socket in a reasonable interval")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Start tmux and have it start the hermetic Elvish shell.
	elvRcPath := filepath.Join(cwd, "website/ttyshot/ttyshot.rc")
	tmuxCmd := exec.Cmd{
		Path: tmuxPath,
		Args: []string{
			"tmux", "-S", tmuxSock, "-f", "website/ttyshot/tmux.rc", "new-session", "-s", "ttyshot",
			"-c", homePath, elvishPath, "-rc", elvRcPath, "-sock", elvSock},
		Stdin:  slave,
		Stdout: slave,
		Stderr: slave,
	}
	go func() {
		// We ignore the Run() error return value because it will normally tell us the tmux exit
		// status was one. We could explicitly test for that error and only call log.Fatal if it was
		// some other error but there really isn't a good reason to do so.
		tmuxCmd.Run()
	}()

	// Capture the output of the Elvish shell.
	captureCmd := exec.Cmd{
		Path:   tmuxPath,
		Args:   []string{"tmux", "-S", tmuxSock, "capture-pane", "-t", "ttyshot", "-p", "-e"},
		Stdin:  devnul,
		Stdout: ttyImage,
		Stderr: os.Stderr,
	}
	go func() {
		<-triggerTtyCapture
		if err := captureCmd.Run(); err != nil {
			log.Fatal(err)
		}
		killTmuxCmd := exec.Cmd{
			Path:   tmuxPath,
			Args:   []string{"tmux", "-S", tmuxSock, "kill-server"},
			Stdin:  devnul,
			Stdout: devnul,
			Stderr: devnul,
		}
		if err := killTmuxCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Killing tmux returned error: %v\n", err)
		}
		ttyCaptureDone <- true
	}()

	return triggerTtyCapture, ttyCaptureDone
}

// Create a hermetic environment for generating a ttyshot. We want to ensure we don't use the real
// home directory, or interactive history, of the person running this tool.
func initEnv() (string, string, func()) {
	// There are systems, such as macOs, which generate a temp dir that includes symlinks in the
	// path. For example, `/var/` => `/private/var`. Expand those symlinks so that Elvish command
	// `tilde-abbr` will behave as expected.
	homePath, err := os.MkdirTemp("", "ttyshot-*")
	if err != nil {
		log.Fatal("unable to create temp home: " + err.Error())
	}
	homePath, err = filepath.EvalSymlinks(homePath)
	if err != nil {
		log.Fatal("unable to resolve symlinks in homePath: " + err.Error())
	}
	// We'll put the Elvish and Tmux socket files in this directory. This makes the "navigation"
	// mode ttyshots a trifle less confusing.
	os.Mkdir(filepath.Join(homePath, "tmp"), 0o700)

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
		log.Fatal("mkdir -p " + dotLocalStateElvish + ": " + err.Error())
	}

	// Copy the Elvish source code to the hermetic home for use in demos of things like Elvish's
	// "navigation" mode.
	copySrcCmd := exec.Cmd{
		Path: "website/tools/cp-elvish.sh",
		Args: []string{"cp-elvish.sh", homePath},
	}
	if err := copySrcCmd.Run(); err != nil {
		log.Fatal(err)
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
			log.Fatal("Unable to remove temp HOME: " + err.Error())
		}
	}

	dbPath := filepath.Join(dotLocalStateElvish, "db.bolt")
	return homePath, dbPath, cleanup
}

func createTtyshot(homePath, dbPath string, script []demoOp, outFile, rawFile *os.File) error {
	master, slave, err := pty.Open()
	if err != nil {
		return err
	}
	winsize := pty.Winsize{Rows: terminalRows, Cols: terminalCols}
	pty.Setsize(master, &winsize)

	// Relay the output of the ttyshot Elvish session to the channel that will capture and evaluate
	// the output; e.g., to detect whether a prompt was seen.
	ttyOutput := make(chan byte, 32*1024)
	go func() {
		for {
			content := make([]byte, 1024)
			n, err := master.Read(content)
			if n == 0 {
				close(ttyOutput)
				return
			}
			if err != nil {
				log.Fatal(err)
			}
			for i := 0; i < n; i++ {
				ttyOutput <- content[i]
			}
		}
	}()

	var ttyImage bytes.Buffer
	triggerTtyCapture, ttyCaptureDone := spawnElvish(homePath, dbPath, slave, &ttyImage)
	trimEmptyLines, err := executeScript(script, master, ttyOutput)
	if err != nil {
		return err
	}

	// Give the ttyshot image a chance to stabilize. Yes, this is not guaranteed to work, but in
	// practice it's rarely needed and even pausing a handful of milliseconds will usually suffice.
	time.Sleep(100 * time.Millisecond)
	triggerTtyCapture <- true
	<-ttyCaptureDone
	// Close the pty master to signal EOF to the processes running inside the simulated terminal.
	// This helps ensure processes running inside the simulated terminal will terminate once we're
	// done capturing the "ttyshot".
	master.Close()

	ttyshot := ttyImage.String()
	rawFile.WriteString(ttyshot)
	// Trim the last, or all, trailing newlines in order to eliminate from the generated HTML
	// unwanted empty lines at the bottom of the ttyshot. The latter behavior occurs if the ttyshot
	// specification includes the `trim-empty` directive.
	if !trimEmptyLines {
		ttyshot = strings.TrimSuffix(ttyshot, "\n")
	} else {
		ttyshot = strings.TrimRight(ttyshot, "\n")
	}
	outFile.WriteString(sgrTextToHTML(ttyshot))
	outFile.WriteString("\n")
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Expected one argument, got %d\n", len(os.Args)-1)
		os.Exit(1)
	}
	specPath := os.Args[1]
	if !strings.HasSuffix(specPath, ".spec") {
		fmt.Fprintf(os.Stderr, "Expected extension \".spec\", found %q\n", path.Ext(specPath))
		os.Exit(2)
	}
	basePath := specPath[:len(specPath)-len(".spec")]
	htmlPath := basePath + ".html"
	rawPath := basePath + ".raw"

	content, err := os.ReadFile(specPath)
	if err != nil {
		log.Fatal(err)
	}

	script, err := parseSpec(content)
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.OpenFile(htmlPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}

	rawFile, err := os.OpenFile(rawPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}

	homePath, dbPath, cleanup := initEnv()
	defer cleanup()
	if err := createTtyshot(homePath, dbPath, script, outFile, rawFile); err != nil {
		log.Fatal(err)
	}
}

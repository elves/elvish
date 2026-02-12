package term

import "fmt"

// OSC 133 shell integration sequences enable terminals to understand shell
// command structure for features like command navigation, intelligent text
// selection, and command status indicators.
//
// Specification:
// https://gitlab.freedesktop.org/Per_Bothner/specifications/blob/master/proposals/semantic-prompts.md
const (
	// OSC133P_R marks a right prompt. This is critical for terminals to
	// correctly classify right prompts vs. user input.
	OSC133P_R = "\033]133;P;k=r\007"

	// OSC133P_S marks a secondary/continuation prompt for multiline input.
	OSC133P_S = "\033]133;P;k=s\007"

	// OSC133B marks the start of user input. This separates prompts from
	// command input, enabling proper command copying and selection.
	OSC133B = "\033]133;B\007"

	// OSC133C marks the start of command execution.
	OSC133C = "\033]133;C\007"
)

// OSC133A returns the sequence that starts a new command and enters prompt
// mode. The following text is assumed to be the initial prompt.
func OSC133A(pid int) string {
	return fmt.Sprintf("\033]133;A;cl=m;aid=%d\007", pid)
}

// OSC133D returns the sequence marking command execution end with exit code.
func OSC133D(exitCode int, pid int) string {
	return fmt.Sprintf("\033]133;D;%d;aid=%d\007", exitCode, pid)
}

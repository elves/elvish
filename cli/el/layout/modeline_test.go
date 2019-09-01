package layout

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/elves/elvish/styled"
)

func TestModePromptNoSpace(t *testing.T) {
	prompt := ModePrompt("TEST", false)()
	wantPrompt := styled.MakeText("TEST", "bold", "lightgray", "bg-magenta")
	if !reflect.DeepEqual(prompt, wantPrompt) {
		fmt.Printf("got prompt %v, want %v", prompt, wantPrompt)
	}
}

func TestModePromptWithSpace(t *testing.T) {
	prompt := ModePrompt("TEST", true)()
	wantPrompt := styled.MakeText("TEST", "bold", "lightgray", "bg-magenta").
		ConcatText(styled.Plain(" "))
	if !reflect.DeepEqual(prompt, wantPrompt) {
		fmt.Printf("got prompt %v, want %v", prompt, wantPrompt)
	}
}

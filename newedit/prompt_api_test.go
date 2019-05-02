package newedit

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/prompt"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestMakePrompt_ElvishVariableLinksToPromptConfig(t *testing.T) {
	ev := eval.NewEvaler()
	// NewEditor calls makePrompt
	ed := NewEditor(devNull, devNull, ev, testStore)
	ev.Global.AddNs("ed", ed.Ns())
	ev.EvalSourceInTTY(eval.NewScriptSource(
		"[t]", "[t]", "ed:prompt = { put 'CUSTOM PROMPT' }"))

	// TODO: Use p.Get() and avoid type assertion
	p := ed.app.Prompt.(*prompt.Prompt)
	content := p.Config().Raw.Compute()

	want := styled.Plain("CUSTOM PROMPT")
	if !reflect.DeepEqual(content, want) {
		t.Errorf("got content %v, want %v", content, want)
	}
}

func TestDefaultPromptForNonRoot(t *testing.T) {
	f := getDefaultPrompt(false)
	wd := util.Getwd()
	testCallPromptStatic(t, f, styled.Text{
		styled.PlainSegment(wd), styled.PlainSegment("> ")})
}

func TestDefaultPromptForRoot(t *testing.T) {
	f := getDefaultPrompt(true)
	wd := util.Getwd()
	testCallPromptStatic(t, f, styled.Text{
		styled.PlainSegment(wd),
		&styled.Segment{styled.Style{Foreground: "red"}, "# "}})
}

func TestDefaultRPrompt(t *testing.T) {
	f := getDefaultRPrompt("elf", "endor")
	testCallPromptStatic(t, f,
		styled.Transform(styled.Plain("elf@endor"), "inverse"))
}

func testCallPromptStatic(t *testing.T, f eval.Callable, want styled.Text) {
	content := callPrompt(&fakeNotifier{}, eval.NewEvaler(), f)
	if !reflect.DeepEqual(content, want) {
		t.Errorf("get prompt result %v, want %v", content, want)
	}
}

func TestCallPrompt_ConvertsValueOutput(t *testing.T) {
	testCallPrompt(t, "put PROMPT", styled.Plain("PROMPT"), false)
	testCallPrompt(t, "styled PROMPT red",
		styled.Transform(styled.Plain("PROMPT"), "red"), false)
}

func TestCallPrompt_ErrorsOnInvalidValueOutput(t *testing.T) {
	testCallPrompt(t, "put good; put [bad]", styled.Plain("good"), true)
}

func TestCallPrompt_ErrorsOnException(t *testing.T) {
	testCallPrompt(t, "fail error", nil, true)
}

func TestCallPrompt_ConvertsBytesOutput(t *testing.T) {
	testCallPrompt(t, "print PROMPT", styled.Plain("PROMPT"), false)
}

func testCallPrompt(t *testing.T, fsrc string, want styled.Text, wantErr bool) {
	ev := eval.NewEvaler()
	ev.EvalSourceInTTY(eval.NewScriptSource(
		"[t]", "[t]", fmt.Sprintf("f = { %s }", fsrc)))
	f := ev.Global["f"].Get().(eval.Callable)
	nt := &fakeNotifier{}

	content := callPrompt(nt, ev, f)
	if !reflect.DeepEqual(content, want) {
		t.Errorf("get prompt result %v, want %v", content, want)
	}

	if wantErr {
		if len(nt.notes) == 0 {
			t.Errorf("got no error, want errors")
		}
	} else {
		if len(nt.notes) > 0 {
			t.Errorf("got errors %v, want none", nt.notes)
		}
	}
}

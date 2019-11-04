package cliedit

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestPrompt(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setup()
	defer cleanup()

	evalf(ev, `edit:prompt = { put '>>> ' }`)
	_, _, stop := start(ed)
	defer stop()
	wantBuf := bb().WritePlain(">>> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}

func TestRPrompt(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setup()
	defer cleanup()

	evalf(ev, `edit:rprompt = { put 'RRR' }`)
	_, _, stop := start(ed)
	defer stop()
	wantBuf := bb().WritePlain("~> ").SetDotToCursor().
		WritePlain(strings.Repeat(" ", testTTYWidth-6) + "RRR").Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}

func TestDefaultPromptForNonRoot(t *testing.T) {
	f := getDefaultPrompt(false)
	testCallPromptStatic(t, f,
		styled.Plain(util.Getwd()).ConcatText(styled.Plain("> ")))
}

func TestDefaultPromptForRoot(t *testing.T) {
	f := getDefaultPrompt(true)
	testCallPromptStatic(t, f,
		styled.Plain(util.Getwd()).ConcatText(styled.MakeText("# ", "red")))
}

func TestDefaultRPrompt(t *testing.T) {
	f := getDefaultRPrompt("elf", "endor")
	testCallPromptStatic(t, f,
		styled.MakeText("elf@endor", "inverse"))
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

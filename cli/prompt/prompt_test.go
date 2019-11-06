package prompt

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/elves/elvish/styled"
)

func TestPrompt_DefaultCompute(t *testing.T) {
	prompt := New(Config{})

	prompt.Trigger(false)
	testUpdate(t, prompt, styled.Plain("???> "))
}

func TestPrompt_ShowsComputedPrompt(t *testing.T) {
	prompt := New(Config{
		Compute: func() styled.Text { return styled.Plain(">>> ") }})

	prompt.Trigger(false)
	testUpdate(t, prompt, styled.Plain(">>> "))
}

func TestPrompt_StalePrompt(t *testing.T) {
	unblockPrompt := make(chan struct{})
	i := 0
	prompt := New(Config{
		Compute: func() styled.Text {
			<-unblockPrompt
			i++
			return styled.Plain(fmt.Sprintf("%d> ", i))
		},
		StaleThreshold: func() time.Duration {
			return 10 * time.Millisecond
		},
	})

	prompt.Trigger(false)
	// The compute function is blocked, so a stale version of the intial
	// "unknown" prompt will be shown.
	testUpdate(t, prompt, styled.MakeText("???> ", "inverse"))

	// The compute function will now return.
	unblockPrompt <- struct{}{}
	// The returned prompt will now be used.
	testUpdate(t, prompt, styled.Plain("1> "))

	// Force a refresh.
	prompt.Trigger(true)
	// The compute function will now be blocked again, so after a while a stale
	// version of the previous prompt will be shown.
	testUpdate(t, prompt, styled.MakeText("1> ", "inverse"))

	// Unblock the compute function.
	unblockPrompt <- struct{}{}
	// The new prompt will now be shown.
	testUpdate(t, prompt, styled.Plain("2> "))
}

func TestPrompt_Eagerness5(t *testing.T) {
	i := 0
	prompt := New(Config{
		Compute: func() styled.Text {
			i++
			return styled.Plain(fmt.Sprintf("%d> ", i))
		},
		Eagerness: func() int { return 5 },
	})

	// An initial update is always triggered.
	prompt.Trigger(false)
	testUpdate(t, prompt, styled.Plain("1> "))

	prompt.Trigger(false)
	testNoUpdate(t, prompt)
}

func testUpdate(t *testing.T, p *Prompt, wantUpdate styled.Text) {
	t.Helper()
	update := <-p.LateUpdates()
	if !reflect.DeepEqual(update, wantUpdate) {
		t.Errorf("got late update %v, want %v", update, wantUpdate)
	}
	current := p.Get()
	if !reflect.DeepEqual(current, wantUpdate) {
		t.Errorf("got current %v, want %v", current, wantUpdate)
	}
}

func testNoUpdate(t *testing.T, p *Prompt) {
	select {
	case update := <-p.LateUpdates():
		t.Errorf("unexpected update %v", update)
	case <-time.After(10 * time.Millisecond):
		// OK
	}
}

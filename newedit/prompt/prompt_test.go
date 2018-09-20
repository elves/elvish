package prompt

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/elves/elvish/styled"
)

func TestPrompt_DeliversResultOnLateUpdates(t *testing.T) {
	content := styled.Unstyled("prompt> ")
	prompt := New(func() styled.Text { return content })
	prompt.Trigger(false)
	update := <-prompt.LateUpdates()
	if !reflect.DeepEqual(update, content) {
		t.Errorf("want update %v, got %v", content, update)
	}

	current := prompt.Get()
	if !reflect.DeepEqual(current, content) {
		t.Errorf("want current %v, got %v", content, current)
	}
}

func TestPrompt_DeliversStaleLastPromptsOnLateUpdates(t *testing.T) {
	unblockPrompt := make(chan struct{})
	content := styled.Unstyled("prompt> ")
	prompt := New(func() styled.Text {
		<-unblockPrompt
		return content
	})
	prompt.Trigger(false)

	update := <-prompt.LateUpdates()
	staleUnknown := styled.Transform(unknownContent, "inverse")
	if !reflect.DeepEqual(update, staleUnknown) {
		t.Errorf("want update %v, got %v", staleUnknown, update)
	}

	unblockPrompt <- struct{}{}

	update = <-prompt.LateUpdates()
	if !reflect.DeepEqual(update, content) {
		t.Errorf("want update %v, got %v", content, update)
	}

	current := prompt.Get()
	if !reflect.DeepEqual(current, content) {
		t.Errorf("want current %v, got %v", content, current)
	}
}

func TestPrompt_DeliversStaleCurrentPromptsOnLateUpdates(t *testing.T) {
	unblockPrompt := make(chan struct{})
	content := styled.Unstyled("prompt> ")
	prompt := New(func() styled.Text {
		<-unblockPrompt
		return content
	})
	prompt.Trigger(false)

	update := <-prompt.LateUpdates()
	staleUnknown := styled.Transform(unknownContent, "inverse")
	if !reflect.DeepEqual(update, staleUnknown) {
		t.Errorf("want update %v, got %v", staleUnknown, update)
	}

	prompt.Trigger(true)
	unblockPrompt <- struct{}{}

	update = <-prompt.LateUpdates()
	wantContent := styled.Transform(content, "inverse")
	if !reflect.DeepEqual(update, wantContent) {
		t.Errorf("want update %v, got %v", wantContent, update)
	}

	unblockPrompt <- struct{}{}

	update = <-prompt.LateUpdates()
	if !reflect.DeepEqual(update, content) {
		t.Errorf("want update %v, got %v", content, update)
	}
}

func TestPrompt_DoesNotRecomputeWhenInSameDir(t *testing.T) {
	i := 0
	prompt := New(func() styled.Text {
		i++
		return styled.Unstyled(strconv.Itoa(i))
	})

	prompt.Trigger(false)

	update := <-prompt.LateUpdates()
	wantPrompt := styled.Unstyled("1")
	if !reflect.DeepEqual(update, wantPrompt) {
		t.Errorf("want update %v, got %v", wantPrompt, update)
	}

	prompt.Trigger(false)

	select {
	case update = <-prompt.LateUpdates():
		t.Errorf("unexpected update %v", update)
	case <-time.After(10 * time.Millisecond):
		// OK
	}
}

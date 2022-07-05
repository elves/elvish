package prompt

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

func TestPrompt_DefaultCompute(t *testing.T) {
	prompt := New(Config{})

	prompt.Trigger(false)
	testUpdate(t, prompt, ui.T("???> "))
}

func TestPrompt_ShowsComputedPrompt(t *testing.T) {
	prompt := New(Config{
		Compute: func() ui.Text { return ui.T(">>> ") }})

	prompt.Trigger(false)
	testUpdate(t, prompt, ui.T(">>> "))
}

func TestPrompt_StalePrompt(t *testing.T) {
	compute, unblock := blockedAutoIncPrompt()
	prompt := New(Config{
		Compute: compute,
		StaleThreshold: func() time.Duration {
			return testutil.Scaled(10 * time.Millisecond)
		},
	})

	prompt.Trigger(true)
	// The compute function is blocked, so a stale version of the initial
	// "unknown" prompt will be shown.
	testUpdate(t, prompt, ui.T("???> ", ui.Inverse))

	// The compute function will now return.
	unblock()
	// The returned prompt will now be used.
	testUpdate(t, prompt, ui.T("1> "))

	// Force a refresh.
	prompt.Trigger(true)
	// The compute function will now be blocked again, so after a while a stale
	// version of the previous prompt will be shown.
	testUpdate(t, prompt, ui.T("1> ", ui.Inverse))

	// Unblock the compute function.
	unblock()
	// The new prompt will now be shown.
	testUpdate(t, prompt, ui.T("2> "))

	// Force a refresh.
	prompt.Trigger(true)
	// Make sure that the compute function is run and stuck.
	testUpdate(t, prompt, ui.T("2> ", ui.Inverse))
	// Queue another two refreshes before the compute function can return.
	prompt.Trigger(true)
	prompt.Trigger(true)
	unblock()
	// Now the new prompt should be marked stale immediately.
	testUpdate(t, prompt, ui.T("3> ", ui.Inverse))
	unblock()
	// However, the two refreshes we requested early only trigger one
	// re-computation, because they are requested while the compute function is
	// stuck, so they can be safely merged.
	testUpdate(t, prompt, ui.T("4> "))
}

func TestPrompt_Eagerness0(t *testing.T) {
	prompt := New(Config{
		Compute:   autoIncPrompt(),
		Eagerness: func() int { return 0 },
	})

	// A forced refresh is always respected.
	prompt.Trigger(true)
	testUpdate(t, prompt, ui.T("1> "))

	// A unforced refresh is not respected.
	prompt.Trigger(false)
	testNoUpdate(t, prompt)

	// No update even if pwd has changed.
	testutil.InTempDir(t)
	prompt.Trigger(false)
	testNoUpdate(t, prompt)

	// Only force updates are respected.
	prompt.Trigger(true)
	testUpdate(t, prompt, ui.T("2> "))
}

func TestPrompt_Eagerness5(t *testing.T) {
	prompt := New(Config{
		Compute:   autoIncPrompt(),
		Eagerness: func() int { return 5 },
	})

	// The initial trigger is respected because there was no previous pwd.
	prompt.Trigger(false)
	testUpdate(t, prompt, ui.T("1> "))

	// No update because the pwd has not changed.
	prompt.Trigger(false)
	testNoUpdate(t, prompt)

	// Update because the pwd has changed.
	testutil.InTempDir(t)
	prompt.Trigger(false)
	testUpdate(t, prompt, ui.T("2> "))
}

func TestPrompt_Eagerness10(t *testing.T) {
	prompt := New(Config{
		Compute:   autoIncPrompt(),
		Eagerness: func() int { return 10 },
	})

	// The initial trigger is respected.
	prompt.Trigger(false)
	testUpdate(t, prompt, ui.T("1> "))

	// Subsequent triggers, force or not, are also respected.
	prompt.Trigger(false)
	testUpdate(t, prompt, ui.T("2> "))
	prompt.Trigger(true)
	testUpdate(t, prompt, ui.T("3> "))
	prompt.Trigger(false)
	testUpdate(t, prompt, ui.T("4> "))
}

func blockedAutoIncPrompt() (func() ui.Text, func()) {
	unblockChan := make(chan struct{})
	i := 0
	compute := func() ui.Text {
		<-unblockChan
		i++
		return ui.T(fmt.Sprintf("%d> ", i))
	}
	unblock := func() {
		unblockChan <- struct{}{}
	}
	return compute, unblock
}

func autoIncPrompt() func() ui.Text {
	i := 0
	return func() ui.Text {
		i++
		return ui.T(fmt.Sprintf("%d> ", i))
	}
}

func testUpdate(t *testing.T, p *Prompt, wantUpdate ui.Text) {
	t.Helper()
	select {
	case <-p.LateUpdates():
		update := p.Get()
		if !reflect.DeepEqual(update, wantUpdate) {
			t.Errorf("got updated %v, want %v", update, wantUpdate)
		}
	case <-time.After(time.Second):
		t.Errorf("no late update after 1 second")
	}
}

func testNoUpdate(t *testing.T, p *Prompt) {
	t.Helper()
	select {
	case update := <-p.LateUpdates():
		t.Errorf("unexpected update %v", update)
	case <-time.After(testutil.Scaled(10 * time.Millisecond)):
		// OK
	}
}

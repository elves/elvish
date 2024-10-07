package edit

import (
	"fmt"
	"os"
	"time"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
)

type promptCfg struct {
	Fn               func() ui.Text
	Eagerness        int
	StaleThreshold   float64
	StaleTransformer func(ui.Text) ui.Text
}

func makeDefaultPromptCfg(fn func() ui.Text) promptCfg {
	return promptCfg{fn, 5, 0.1, func(t ui.Text) ui.Text { return ui.StyleText(t, ui.Inverse) }}
}

type promptKey struct {
	Fn        string
	Eagerness int
	Wd        string
	Buffer    string
}

var promptMaxBlock = 10 * time.Millisecond

// Call p.Fn and use the result to set key in c.
//
//   - If p.Fn returns within [promptMaxBlock], blocks until it returns and
//     sets key synchronously.
//   - If p.Fn takes longer than [promptMaxBlock] but shorter than
//     [p.StaleThreshold], returns after [promptMaxBlock], and sets key
//     asynchronously when p.Fn returns.
//   - If p.Fn takes more than [p.StaleThreshold], returns after
//     [promptMaxBlock], transforms the previous prompt value using
//     [p.StaleTransformer] after [p.StaleThreshold], and sets key asynchronously
//     when [p.Fn] returns.
func callPrompt(c etk.Context, key string, p promptCfg, buf string) {
	var keyChanged bool
	etk.BindState(c, key+"-key", promptKey{}).Swap(
		func(oldKey promptKey) promptKey {
			key := promptKey{Fn: fmt.Sprintf("%p", p.Fn), Eagerness: p.Eagerness}
			if p.Eagerness >= 5 {
				var err error
				key.Wd, err = os.Getwd()
				if err != nil {
					key.Wd = "error"
				}
			}
			if p.Eagerness >= 10 {
				key.Buffer = buf
			}
			keyChanged = oldKey != key
			return key
		})
	if !keyChanged {
		return
	}

	promptVar := etk.BindState(c, key, ui.Text(nil))

	promptCh, ready := withTimeout(promptMaxBlock, p.Fn)
	if ready {
		promptVar.Set(<-promptCh)
		return
	}
	go func() {
		staleThreshold := time.Duration(p.StaleThreshold * float64(time.Second))
		select {
		case prompt := <-promptCh:
			// TODO: Check epoch
			promptVar.Set(prompt)
			c.Refresh()
		case <-time.After(staleThreshold - promptMaxBlock):
			// TODO: Check epoch
			promptVar.Swap(func(t ui.Text) ui.Text {
				return p.StaleTransformer(t)
			})
			c.Refresh()
			prompt := <-promptCh
			// TODO: Check epoch
			promptVar.Set(prompt)
			c.Refresh()
		}
	}()
}

func withTimeout[T any](d time.Duration, f func() T) (<-chan T, bool) {
	ch := make(chan T, 1)
	go func() { ch <- f() }()
	select {
	case ret := <-ch:
		ch <- ret
		return ch, true
	case <-time.After(d):
		return ch, false
	}
}

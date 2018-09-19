// Package prompt provides an implementation of the core.Prompt interface.
package prompt

import (
	"os"
	"sync"
	"time"

	"github.com/elves/elvish/styled"
)

type prompt struct {
	config Config

	// Working directory when prompt was last updated.
	lastWd string
	// Channel for update requests.
	updateReq chan struct{}
	// Channel on which prompt contents are delivered.
	ch chan styled.Text
	// Last computed prompt content.
	last styled.Text
	// Mutex for guarding access to the last field.
	lastMutex sync.RWMutex
}

var unknownContent = styled.Unstyled("???> ")

func defaultStaleTransform(t styled.Text) styled.Text {
	return styled.Transform(t, "inverse")
}

func makePrompt(fn func() styled.Text) *prompt {
	p := &prompt{
		Config{Raw: RawConfig{
			fn, defaultStaleTransform, 2 * time.Millisecond, 5}},
		"", make(chan struct{}, 1), make(chan styled.Text, 1),
		unknownContent, sync.RWMutex{}}
	// TODO: Don't keep a goroutine running.
	go p.loop()
	return p
}

func (p *prompt) loop() {
	content := unknownContent
	ch := make(chan styled.Text)
	for range p.updateReq {
		go func() {
			ch <- p.config.Compute()()
		}()

		select {
		case <-time.After(p.config.StaleThreshold()):
			// The prompt callback did not finish within the threshold. Send the
			// previous content, marked as stale.
			p.send(p.config.StaleTransform()(content))
			content = <-ch

			select {
			case <-p.updateReq:
				// If another update is already requested by the time we finish,
				// keep marking the prompt as stale. This reduces flickering.
				p.send(p.config.StaleTransform()(content))
				p.queueUpdate()
			default:
				p.send(content)
			}
		case content = <-ch:
			p.send(content)
		}
	}
}

func (p *prompt) Trigger(force bool) {
	if force || p.shouldUpdate() {
		p.queueUpdate()
	}
}

func (p *prompt) Get() styled.Text {
	p.lastMutex.RLock()
	defer p.lastMutex.RUnlock()
	return p.last
}

func (p *prompt) LateUpdates() <-chan styled.Text {
	return p.ch
}

func (p *prompt) queueUpdate() {
	select {
	case p.updateReq <- struct{}{}:
	default:
	}
}

func (p *prompt) send(content styled.Text) {
	p.lastMutex.Lock()
	p.last = content
	p.lastMutex.Unlock()
	p.ch <- content
}

func (p *prompt) shouldUpdate() bool {
	eagerness := p.config.Eagerness()
	if eagerness >= 10 {
		return true
	}
	if eagerness >= 5 {
		wd, err := os.Getwd()
		if err != nil {
			wd = "error"
		}
		oldWd := p.lastWd
		p.lastWd = wd
		return wd != oldWd
	}
	return false
}

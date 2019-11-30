// Package prompt provides an implementation of the cli.Prompt interface.
package prompt

import (
	"os"
	"sync"
	"time"

	"github.com/elves/elvish/ui"
)

// Prompt implements a prompt that is executed asynchronously.
type Prompt struct {
	config Config

	// Working directory when prompt was last updated.
	lastWd string
	// Channel for update requests.
	updateReq chan struct{}
	// Channel on which prompt contents are delivered.
	ch chan ui.Text
	// Last computed prompt content.
	last ui.Text
	// Mutex for guarding access to the last field.
	lastMutex sync.RWMutex
}

// Config keeps configurations for the prompt.
type Config struct {
	// The function that computes the prompt.
	Compute func() ui.Text
	// Function to transform stale prompts.
	StaleTransform func(ui.Text) ui.Text
	// Threshold for a prompt to be considered as stale.
	StaleThreshold func() time.Duration
	// How eager the prompt should be updated. When >= 5, updated when directory
	// is changed. When >= 10, always update. Default is 5.
	Eagerness func() int
}

func defaultStaleTransform(t ui.Text) ui.Text {
	return ui.TransformText(t, "inverse")
}

const defaultStaleThreshold = 200 * time.Millisecond

const defaultEagerness = 5

var unknownContent = ui.PlainText("???> ")

// New makes a new prompt.
func New(cfg Config) *Prompt {
	if cfg.Compute == nil {
		cfg.Compute = func() ui.Text { return unknownContent }
	}
	if cfg.StaleTransform == nil {
		cfg.StaleTransform = defaultStaleTransform
	}
	if cfg.StaleThreshold == nil {
		cfg.StaleThreshold = func() time.Duration { return defaultStaleThreshold }
	}
	if cfg.Eagerness == nil {
		cfg.Eagerness = func() int { return defaultEagerness }
	}
	p := &Prompt{
		cfg,
		"", make(chan struct{}, 1), make(chan ui.Text, 1),
		unknownContent, sync.RWMutex{}}
	// TODO: Don't keep a goroutine running.
	go p.loop()
	return p
}

func (p *Prompt) loop() {
	content := unknownContent
	ch := make(chan ui.Text)
	for range p.updateReq {
		go func() {
			ch <- p.config.Compute()
		}()

		select {
		case <-time.After(p.config.StaleThreshold()):
			// The prompt callback did not finish within the threshold. Send the
			// previous content, marked as stale.
			p.send(p.config.StaleTransform(content))
			content = <-ch

			select {
			case <-p.updateReq:
				// If another update is already requested by the time we finish,
				// keep marking the prompt as stale. This reduces flickering.
				p.send(p.config.StaleTransform(content))
				p.queueUpdate()
			default:
				p.send(content)
			}
		case content = <-ch:
			p.send(content)
		}
	}
}

// Trigger triggers an update to the prompt.
func (p *Prompt) Trigger(force bool) {
	if force || p.shouldUpdate() {
		p.queueUpdate()
	}
}

// Get returns the current content of the prompt.
func (p *Prompt) Get() ui.Text {
	p.lastMutex.RLock()
	defer p.lastMutex.RUnlock()
	return p.last
}

// LateUpdates returns a channel on which late updates are made available.
func (p *Prompt) LateUpdates() <-chan ui.Text {
	return p.ch
}

func (p *Prompt) queueUpdate() {
	select {
	case p.updateReq <- struct{}{}:
	default:
	}
}

func (p *Prompt) send(content ui.Text) {
	p.lastMutex.Lock()
	p.last = content
	p.lastMutex.Unlock()
	p.ch <- content
}

func (p *Prompt) shouldUpdate() bool {
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

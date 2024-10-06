package main

import (
	"fmt"
	"time"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

func Timer(c etk.Context) (etk.View, etk.React) {
	startTimeVar := etk.State(c, "start-time", time.Now())
	durationView, durationReact := c.Subcomp("duration", comps.TextArea)
	durationBufferVar := etk.BindState(c, "duration/buffer", comps.TextBuffer{})
	resetView, resetReact := c.Subcomp("reset",
		etk.WithInit(Button, "label", "Reset", "submit", func() {
			startTimeVar.Set(time.Now())
		}))
	formView, formReact := Form(c,
		FormComp{"Duration: ", durationView, durationReact, false},
		FormComp{"", resetView, resetReact, false},
	)

	elapsed := time.Since(startTimeVar.Get())
	if d, err := time.ParseDuration(durationBufferVar.Get().Content); err == nil {
		elapsed = min(elapsed, d)
		// TODO: Remember previous duration if invalid?
	}
	elapsedText := ui.T(fmt.Sprintf(
		"  Elapsed time: %.1fs", float64(elapsed)/float64(time.Second)))

	if launchedVar := etk.State(c, "launched", false); !launchedVar.Get() {
		launchedVar.Set(true)
		go func() {
			for !c.Finished() {
				time.Sleep(time.Second / 10)
				c.Refresh()
			}
		}()
	}

	// TODO: Progress bar
	return etk.Box(`
		elapsed=
		[form=]`, etk.Text(elapsedText), formView), formReact
}

package edit

import (
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
)

func defaultHistoryProvider(hs *histStore, code string) string {
	if hs == nil {
		return ""
	}
	if code == "" {
		return ""
	}

	// Get commands from history that start with the current input
	// TODO: Optimize this to avoid fetching all commands
	cmds, err := hs.AllCmds()
	if err != nil {
		return ""
	}

	// Search backwards through history for the most recent match
	for i := len(cmds) - 1; i >= 0; i-- {
		text := cmds[i].Text
		// Match commands that start with current input
		// Also handle multiline commands - only match against the first line
		if strings.HasPrefix(text, code) && len(text) > len(code) {
			text = TrimSpaceRight(text)
			// Return the remaining part after the current input
			return text[len(code):]
		}
	}

	return ""
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func TrimSpaceRight(s string) string {
	start := 0
	stop := len(s)
	for ; stop > start; stop-- {
		c := s[stop-1]
		if c >= utf8.RuneSelf {
			// start has been already trimmed above, should trim end only
			return strings.TrimRightFunc(s[start:stop], unicode.IsSpace)
		}
		if asciiSpace[c] == 0 {
			break
		}
	}

	return s[start:stop]
}

func initAutoSuggestionSpec(appSpec *cli.AppSpec, hs *histStore, enabled vars.PtrVar, provider vars.PtrVar, ev *eval.Evaler) {
	appSpec.AutoSuggestionProvider = func(code string) string {
		if !enabled.GetRaw().(bool) {
			return ""
		}

		// Check if user has set a custom provider function
		providerFn := provider.GetRaw()
		if fn, ok := providerFn.(eval.Callable); ok {
			// Call user-defined provider function
			var result string
			valuesCb := func(ch <-chan any) {
				for v := range ch {
					if s, ok := v.(string); ok && result == "" {
						result = s
						return
					}
				}
			}
			bytesCb := func(_ *os.File) {}

			port, done, err := eval.PipePort(valuesCb, bytesCb)
			if err != nil {
				return ""
			}
			err = ev.Call(fn,
				eval.CallCfg{Args: []any{code}, From: "[auto-suggestion provider]"},
				eval.EvalCfg{Ports: []*eval.Port{nil, port, nil}})
			done()
			if err == nil && result != "" {
				return result
			}
			return ""
		}

		// Fall back to default history-based provider
		return defaultHistoryProvider(hs, code)
	}
}

func initAutoSuggestionAPI(app cli.App, enabled vars.PtrVar, provider vars.PtrVar, nb eval.NsBuilder) {
	acceptFn := func() {
		codeArea, ok := focusedCodeArea(app)
		if !ok {
			return
		}
		codeArea.MutateState(func(s *tk.CodeAreaState) {
			// Only accept if there's an autosuggestion pending
			if s.Pending.Content != "" && s.Pending.AutoSuggestion {
				s.ApplyPending()
			}
		})
	}

	nb.AddNs("auto-suggestion",
		eval.BuildNsNamed("edit:auto-suggestion").
			AddVar("enabled", enabled).
			AddVar("provider", provider).
			AddGoFn("accept", acceptFn))
}

// Package etktest provides facilities for testing Etk components.
package etktest

import (
	"fmt"
	"strings"
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/wcwidth"
)

type sendOpts struct{ ShowReaction bool }

func (*sendOpts) SetDefaultOptions() {}

type renderOpts struct{ Width, Height int }

func (opts *renderOpts) SetDefaultOptions() {
	opts.Width = 40
	opts.Height = 10
}

// Setup adds the following commands to the global namespace that manipulates
// the component:
//
//   - setup: recreate the component with initial state values.
//   - send: send events.
//   - render: render in a Styledown-like format.
//   - refresh: refresh the component.
func Setup(t *testing.T, ev *eval.Evaler, f etk.Comp) {
	sc := etk.Stateful(ev.CallFrame("etktest"), f)
	// We can't pass w.Finish here because w might be reassigned by the setup
	// function below.
	t.Cleanup(func() { sc.Finish() })
	// A note on the nature of the test fixture:
	//
	// An important difference between the event loop, implemented by [etk.Run]
	// and used in the real REPL, and this test fixture is that the former is a
	// continuously running push system, while the latter is an on-demand pull
	// system. For example, the user could type "x" at the real REPL at any
	// time, and the main loop reacts to that and refreshes the UI immediately.
	// On the other hand, emulating the same event in the test fixture with the
	// "send" command mutates some internal state, but the UI change can't be
	// observed until the next invocation of the "render" command.
	//
	// This difference results in some subtle logic to handle refresh requests.
	// Because the test fixture is invoked on demand, it has no way of handling
	// refresh requests in real time. At the same time, the fact that these
	// requests have not been handled can't be observed until a call to the
	// "send" or "render" command either, so we overcome that by handling them
	// at the start of both commands.
	ev.ExtendGlobal(eval.BuildNs().AddGoFns(map[string]any{
		"setup": func(m vals.Map) {
			sc.Finish()
			sc = etk.Stateful(ev.CallFrame("etktest"),
				etk.WithInit(f, must.OK1(convertSetStates(m))...))
		},
		"send": func(fm *eval.Frame, opts sendOpts, args ...any) error {
			events, err := parseEvents(args)
			if err != nil {
				return err
			}
			for _, ev := range events {
				sc.RefreshIfRequested()
				reaction := sc.React(ev)
				if opts.ShowReaction {
					fmt.Fprintln(fm.ByteOutput(), reaction)
				}
			}
			return nil
		},
		"render": func(fm *eval.Frame, opts renderOpts) error {
			sc.RefreshIfRequested()
			buf := sc.Render(opts.Width, opts.Height)
			sd, err := bufferToStyleDown(buf, globalStylesheet)
			if err != nil {
				return err
			}
			_, err = fm.ByteOutput().WriteString(sd)
			return err
		},
		"wait-refresh": sc.WaitRefresh,
		"refresh":      sc.Refresh,
	}).Ns())
}

func MakeFixture(f etk.Comp) func(*testing.T, *eval.Evaler) {
	return func(t *testing.T, ev *eval.Evaler) { Setup(t, ev, f) }
}

func parseEvents(args []any) ([]term.Event, error) {
	var events []term.Event
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			for _, r := range arg {
				events = append(events, term.KeyEvent{Rune: r})
			}
		case vals.List:
			for it := arg.Iterator(); it.HasElem(); it.Next() {
				elem := it.Elem()
				switch elem := elem.(type) {
				case string:
					switch elem {
					case "start-paste":
						events = append(events, term.PasteSetting(true))
					case "end-paste":
						events = append(events, term.PasteSetting(false))
					case "mouse-dummy":
						events = append(events, term.MouseEvent{})
					default:
						key, err := ui.ParseKey(elem)
						if err != nil {
							return nil, err
						}
						events = append(events, term.KeyEvent(key))
					}
				default:
					return nil, fmt.Errorf("element of list argument must be string, got %s", vals.ReprPlain(elem))
				}
			}
		default:
			return nil, fmt.Errorf("argument must be string or list, got %s", vals.ReprPlain(arg))
		}
	}
	return events, nil
}

// TODO: This duplicates part of styledown pkg.
var builtinStyleDownChars = map[ui.Style]rune{
	{}:                 ' ',
	{Bold: true}:       '*',
	{Underlined: true}: '_',
	{Inverse: true}:    '#',
	{Fg: ui.Red}:       'R',
	{Fg: ui.Green}:     'G',
	{Fg: ui.Magenta}:   'M',
}

// TODO: This duplicates much of (*term.Buffer).TTYString.
func bufferToStyleDown(b *term.Buffer, ss stylesheet) (string, error) {
	var sb strings.Builder
	// Top border
	sb.WriteString("┌" + strings.Repeat("─", b.Width) + "┐\n")
	for i, line := range b.Lines {
		// Write the content line.
		sb.WriteRune('│')
		usedWidth := 0
		for _, cell := range line {
			sb.WriteString(cell.Text)
			usedWidth += wcwidth.Of(cell.Text)
		}
		var rightPadding string
		if usedWidth < b.Width {
			rightPadding = strings.Repeat(" ", b.Width-usedWidth)
			sb.WriteString(rightPadding)
		}
		sb.WriteString("│\n")

		// Write the style line.
		// TODO: I shouldn't have to keep track of the column number manually
		sb.WriteRune('│')
		col := 0
		for _, cell := range line {
			style := ui.StyleFromSGR(cell.Style)
			var styleChar rune
			if char, ok := builtinStyleDownChars[style]; ok {
				styleChar = char
			} else if char, ok := ss.charForStyle[style]; ok {
				styleChar = char
			} else {
				return "", fmt.Errorf("no char for style: %v", style)
			}
			styleStr := string(styleChar)
			if i == b.Dot.Line && col == b.Dot.Col {
				styleStr += "\u0305\u0302" // combining overline + combining circumflex
			}
			sb.WriteString(strings.Repeat(styleStr, wcwidth.Of(cell.Text)))
			col += wcwidth.Of(cell.Text)
		}
		if i == b.Dot.Line && col <= b.Dot.Col {
			sb.WriteString(strings.Repeat(" ", b.Dot.Col-col+1))
			sb.WriteString("\u0305\u0302")
			sb.WriteString(strings.Repeat(" ", b.Width-b.Dot.Col-1))
		} else {
			sb.WriteString(rightPadding)
		}
		sb.WriteString("│\n")
	}
	// Bottom border
	sb.WriteString("└" + strings.Repeat("─", b.Width) + "┘\n")

	return sb.String(), nil
}

var globalStylesheet = newStylesheet(map[rune]string{
	'r': "red",
})

type stylesheet struct {
	stringStyling map[rune]string
	charForStyle  map[ui.Style]rune
}

func newStylesheet(stringStyling map[rune]string) stylesheet {
	charForStyle := make(map[ui.Style]rune)
	for r, s := range stringStyling {
		var st ui.Style
		ui.ApplyStyling(st, ui.ParseStyling(s))
		charForStyle[st] = r
	}
	return stylesheet{stringStyling, charForStyle}
}

// Same as convertSetStates from pkg/etk, copied to avoid the need to export it.
func convertSetStates(m vals.Map) ([]any, error) {
	var setStates []any
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		name, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("key should be string")
		}
		setStates = append(setStates, name, v)
	}
	return setStates, nil
}

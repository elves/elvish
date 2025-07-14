package comps

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

// TextArea stores text and supports editing it.
//
// State variables:
//   - prompt and rprompt: both [ui.Text]
//   - buffer: a [TextBuffer]
//   - pending: a [PendingText]
//   - highlighter: a function taking a string and returning a [ui.Text]
//     (the highlighted text) and a slice of [ui.Text] (any "tips")
func TextArea(c etk.Context) (etk.View, etk.React) {
	quotePasteVar := etk.State(c, "quote-paste", false)

	pastingVar := etk.State(c, "pasting", false)
	pasteBufferVar := etk.State(c, "paste-buffer", &strings.Builder{})
	innerView, innerReact := textAreaWithAbbr(c)
	bufferVar := etk.BindState(c, "buffer", TextBuffer{})

	// TODO: default key binding?
	binding := c.BindingNopDefault()
	return innerView, func(event term.Event) etk.Reaction {
		if r := binding(event); r != etk.Unused {
			return r
		}
		switch event := event.(type) {
		case term.PasteSetting:
			startPaste := bool(event)
			// TODO:
			// resetInserts()
			if startPaste {
				pastingVar.Set(true)
			} else {
				text := pasteBufferVar.Get().String()
				pasteBufferVar.Set(new(strings.Builder))
				pastingVar.Set(false)

				if quotePasteVar.Get() {
					text = parse.Quote(text)
				}
				bufferVar.Swap(insertAtDot(text))
			}
			return etk.Consumed
		case term.KeyEvent:
			key := ui.Key(event)
			if pastingVar.Get() {
				if isFuncKey(key) {
					// TODO: Notify the user of the error, or insert the
					// original character as is.
				} else {
					pasteBufferVar.Get().WriteRune(key.Rune)
				}
				return etk.Consumed
			}
		}
		return innerReact(event)
	}
}

func textAreaWithAbbr(c etk.Context) (etk.View, etk.React) {
	abbrVar := etk.State(c, "abbr", func(func(a, f string)) {})
	cmdAbbrVar := etk.State(c, "command-abbr", func(func(a, f string)) {})
	smallWordAbbr := etk.State(c, "small-word-abbr", func(func(a, f string)) {})

	streakVar := etk.State(c, "streak", "")
	innerView, innerReact := textAreaCore(c)
	bufferVar := etk.BindState(c, "buffer", TextBuffer{})
	return innerView, func(event term.Event) etk.Reaction {
		if keyEvent, ok := event.(term.KeyEvent); ok {
			bufferBefore := bufferVar.Get()
			reaction := innerReact(event)
			if reaction != etk.Consumed {
				return reaction
			}
			buffer := bufferVar.Get()
			if inserted, ok := isLiteralInsert(keyEvent, bufferBefore, buffer); ok {
				streak := streakVar.Get() + inserted
				if newBuffer, ok := expandSimpleAbbr(abbrVar.Get(), buffer, streak); ok {
					bufferVar.Set(newBuffer)
					streakVar.Set("")
					return etk.Consumed
				}
				if newBuffer, ok := expandCmdAbbr(cmdAbbrVar.Get(), buffer); ok {
					bufferVar.Set(newBuffer)
					streakVar.Set("")
					return etk.Consumed
				}
				if newBuffer, ok := expandSmallWordAbbr(smallWordAbbr.Get(), buffer, streak, keyEvent.Rune, categorizeSmallWord); ok {
					bufferVar.Set(newBuffer)
					streakVar.Set("")
					return etk.Consumed
				}
				streakVar.Set(streak)
			} else {
				streakVar.Set("")
			}
			return etk.Consumed
		} else {
			return innerReact(event)
		}
	}
}

func isLiteralInsert(event term.KeyEvent, before, after TextBuffer) (string, bool) {
	key := ui.Key(event)
	if isFuncKey(key) {
		return "", false
	} else {
		text := string(key.Rune)
		if after == insertAtDot(text)(before) {
			return text, true
		} else {
			return "", false
		}
	}
}

func textAreaCore(c etk.Context) (etk.View, etk.React) {
	promptVar := etk.State(c, "prompt", ui.T(""))
	rpromptVar := etk.State(c, "rprompt", ui.T(""))
	bufferVar := etk.State(c, "buffer", TextBuffer{})
	pendingVar := etk.State(c, "pending", PendingText{})
	highlighterVar := etk.State(c, "highlighter",
		func(code string) (ui.Text, []ui.Text) { return ui.T(code), nil })

	buffer := bufferVar.Get()
	code, pFrom, pTo := patchPending(buffer, pendingVar.Get())
	styledCode, tips := highlighterVar.Get()(code.Content)
	if pFrom < pTo {
		// Apply stylingForPending to [pFrom, pTo)
		parts := styledCode.Partition(pFrom, pTo)
		pending := ui.StyleText(parts[1], stylingForPending)
		styledCode = ui.Concat(parts[0], pending, parts[2])
	}

	view := &textAreaView{
		promptVar.Get(), rpromptVar.Get(),
		styledCode, bufferVar.Get().Dot, tips,
	}
	return view, func(event term.Event) etk.Reaction {
		if event, ok := event.(term.KeyEvent); ok {
			key := ui.Key(event)
			// Implement the absolute essential functionalities here. Others
			// can be added via keybindings.
			switch key {
			case ui.K(ui.Backspace), ui.K('H', ui.Ctrl):
				bufferVar.Swap(backspace)
				return etk.Consumed
			case ui.K(ui.Enter):
				return etk.Finish
			default:
				if key == ui.K(ui.Enter, ui.Alt) || (!isFuncKey(key) && unicode.IsGraphic(key.Rune)) {
					bufferVar.Swap(insertAtDot(string(key.Rune)))
					return etk.Consumed
				}
			}
		}
		return etk.Unused
	}
}

// TextBuffer represents the buffer of a [TextArea].
type TextBuffer struct {
	// Content of the buffer.
	Content string
	// Position of the dot (more commonly known as the cursor), as a byte index
	// into Content.
	Dot int
}

func insertAtDot(text string) func(TextBuffer) TextBuffer {
	return func(buf TextBuffer) TextBuffer {
		return TextBuffer{
			Content: buf.Content[:buf.Dot] + text + buf.Content[buf.Dot:],
			Dot:     buf.Dot + len(text),
		}
	}
}

func backspace(buf TextBuffer) TextBuffer {
	_, chop := utf8.DecodeLastRuneInString(buf.Content[:buf.Dot])
	return TextBuffer{
		Content: buf.Content[:buf.Dot-chop] + buf.Content[buf.Dot:],
		Dot:     buf.Dot - chop,
	}
}

// PendingText represents pending text, such as during completion.
type PendingText struct {
	// Beginning index of the text area that the pending code replaces, as a
	// byte index into RawState.Code.
	From int
	// End index of the text area that the pending code replaces, as a byte
	// index into RawState.Code.
	To int
	// The content of the pending code.
	Content string
}

func isFuncKey(key ui.Key) bool {
	return key.Mod != 0 || key.Rune < 0
}

// Duplicate with pkg/cli/tk/codearea_render.go

func PatchPending(buf TextBuffer, p PendingText) (TextBuffer, int, int) {
	if p.From > p.To || p.From < 0 || p.To > len(buf.Content) {
		// Invalid Pending.
		return buf, 0, 0
	}
	if p.From == p.To && p.Content == "" {
		return buf, 0, 0
	}
	newContent := buf.Content[:p.From] + p.Content + buf.Content[p.To:]
	newDot := 0
	switch {
	case buf.Dot < p.From:
		// Dot is before the replaced region. Keep it.
		newDot = buf.Dot
	case buf.Dot >= p.From && buf.Dot < p.To:
		// Dot is within the replaced region. Place the dot at the end.
		newDot = p.From + len(p.Content)
	case buf.Dot >= p.To:
		// Dot is after the replaced region. Maintain the relative position of
		// the dot.
		newDot = buf.Dot - (p.To - p.From) + len(p.Content)
	}
	return TextBuffer{Content: newContent, Dot: newDot}, p.From, p.From + len(p.Content)
}

// TODO: there should also be an ApplyPending function taking an etk.Context,
// and there should be a public API to build sub-contexts because we need to
// call the function with the TextBuffer's context instead of the global one.

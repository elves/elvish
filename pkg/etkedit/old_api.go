package edit

import (
	"errors"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/eval/vals"
)

var errEditorNotActive = errors.New("editor is not active")

func (ed *Editor) codeBufferBuiltins() map[string]any {
	m := make(map[string]any, len(bufferBuiltinsData))
	for name, fn := range bufferBuiltinsData {
		m[name] = wrapCtxFn(ed, func(c etk.Context) {
			bufferVar(c).Swap(func(buf comps.TextBuffer) comps.TextBuffer {
				fn(&buf)
				return buf
			})
		})
	}
	return m
}

// Used to implement $edit:current-command and $edit:-dot.
type bufferFieldVar[T any] struct {
	ed  *Editor
	get func(comps.TextBuffer) T
	set func(comps.TextBuffer, T) comps.TextBuffer
}

func (bv bufferFieldVar[T]) Get() any {
	c, err := etkCtx(bv.ed)
	if err != nil {
		var zero T
		return zero
	}
	return bv.get(bufferVar(c).Get())
}

func (bv bufferFieldVar[T]) Set(anyVal any) error {
	c, err := etkCtx(bv.ed)
	if err != nil {
		return err
	}
	var val T
	err = vals.ScanToGo(anyVal, &val)
	if err != nil {
		return err
	}
	bufferVar(c).Swap(func(buf comps.TextBuffer) comps.TextBuffer {
		return bv.set(buf, val)
	})
	return nil
}

// Used to implement builtin functions that expect a Context.
func wrapCtxFn(ed *Editor, f func(etk.Context)) func() error {
	return func() error {
		c, err := etkCtx(ed)
		if err != nil {
			return err
		}
		f(c)
		return nil
	}
}

// Like wrapCtxFn, but f also gets access to the editor itself.
func wrapCtxFnEd(ed *Editor, f func(*Editor, etk.Context)) func() error {
	return func() error {
		c, err := etkCtx(ed)
		if err != nil {
			return err
		}
		f(ed, c)
		return nil
	}
}

// Used to implement builtin functions that expect a Context and returns an error.
/*
func wrapCtxFnErr(ed *Editor, f func(etk.Context) error) func() error {
	return func() error {
		c, err := etkCtx(ed)
		if err != nil {
			return err
		}
		return f(c)
	}
}
*/

// Like wrapCtxFn, but f takes an additional argument.
func wrapCtxFn1[T any](ed *Editor, f func(etk.Context, T)) func(T) error {
	return func(a T) error {
		c, err := etkCtx(ed)
		if err != nil {
			return err
		}
		f(c, a)
		return nil
	}
}

// Implements edit:insert-at-dot.
func insertAtDot(c etk.Context, text string) {
	bufferVar(c).Swap(insertAtDotSwapper(text))
}

func insertAtDotSwapper(text string) func(comps.TextBuffer) comps.TextBuffer {
	return func(buf comps.TextBuffer) comps.TextBuffer {
		return comps.TextBuffer{
			Content: buf.Content[:buf.Dot] + text + buf.Content[buf.Dot:],
			Dot:     buf.Dot + len(text),
		}
	}
}

// Implements edit:replace-input.
func replaceInput(c etk.Context, text string) {
	bufferVar(c).Set(comps.TextBuffer{Content: text, Dot: len(text)})
}

func bufferVar(c etk.Context) etk.StateVar[comps.TextBuffer] {
	return etk.BindState(c, "code/buffer", comps.TextBuffer{})
}

func etkCtx(ed *Editor) (etk.Context, error) {
	ctxPtr := getField(ed, &ed.etkCtx)
	if ctxPtr == nil {
		return etk.Context{}, errEditorNotActive
	}
	return *ctxPtr, nil
}

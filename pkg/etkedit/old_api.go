package edit

import (
	"errors"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/eval/vals"
)

var errEditorNotActive = errors.New("editor is not active")

func (ed *Editor) wrapCtxFn(f func(etk.Context) error) func() error {
	return func() error {
		ctxPtr := getField(ed, &ed.etkCtx)
		if ctxPtr == nil {
			return errEditorNotActive
		}
		return f(*ctxPtr)
	}
}

func (ed *Editor) codeBufferBuiltins() map[string]any {
	m := make(map[string]any, len(bufferBuiltinsData))
	for name, fn := range bufferBuiltinsData {
		m[name] = ed.wrapCtxFn(func(c etk.Context) error {
			bufferVar := etk.BindState(c, "code/buffer", comps.TextBuffer{})
			bufferVar.Swap(func(buf comps.TextBuffer) comps.TextBuffer {
				fn(&buf)
				return buf
			})
			return nil
		})
	}
	return m
}

type bufferVar[T any] struct {
	ed  *Editor
	get func(comps.TextBuffer) T
	set func(comps.TextBuffer, T) comps.TextBuffer
}

func (bv bufferVar[T]) Get() any {
	bufferVar, ok := getBufferVar(bv.ed)
	if !ok {
		var zero T
		return zero
	}
	return bv.get(bufferVar.Get())
}

func (bv bufferVar[T]) Set(anyVal any) error {
	bufferVar, ok := getBufferVar(bv.ed)
	if !ok {
		return errEditorNotActive
	}
	var val T
	err := vals.ScanToGo(anyVal, &val)
	if err != nil {
		return err
	}
	bufferVar.Swap(func(buf comps.TextBuffer) comps.TextBuffer {
		return bv.set(buf, val)
	})
	return nil
}

func getBufferVar(ed *Editor) (etk.StateVar[comps.TextBuffer], bool) {
	ctxPtr := getField(ed, &ed.etkCtx)
	if ctxPtr == nil {
		return etk.StateVar[comps.TextBuffer]{}, false
	}
	return etk.BindState(*ctxPtr, "code/buffer", comps.TextBuffer{}), true
}

/*
func (ed *Editor) wrapBufferFn(f func(comps.TextBuffer) comps.TextBuffer) func() error {
	return ed.wrapCtxFn(wrapBufferFn(f))
}

func wrapBufferFn(f func(comps.TextBuffer) comps.TextBuffer) func(etk.Context) error {
	return func(c etk.Context) error {
		etk.BindState(c, "code/buffer", comps.TextBuffer{}).Swap(f)
		return nil
	}
}
*/

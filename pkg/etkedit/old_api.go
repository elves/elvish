package edit

import (
	"errors"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
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

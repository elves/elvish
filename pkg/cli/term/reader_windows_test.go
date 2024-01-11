package term

import (
	"testing"
	"unicode/utf16"

	"src.elv.sh/pkg/sys/ewindows"
	"src.elv.sh/pkg/tt"
	"src.elv.sh/pkg/ui"
)

func TestConvertEvent(t *testing.T) {
	r1, r2 := utf16.EncodeRune('😀')

	tt.Test(t, convertEvent,
		// Only convert KeyEvent
		Args(&ewindows.MouseEvent{}).Rets(nil),
		// Only convert KeyDown events
		Args(&ewindows.KeyEvent{BKeyDown: 0}).Rets(nil),

		Args(charKeyEvent('a', 0)).Rets(K('a')),
		Args(charKeyEvent('A', shift)).Rets(K('A')),
		Args(charKeyEvent('µ', leftCtrl|rightAlt)).Rets(K('µ')),
		Args(charKeyEvent('ẞ', leftCtrl|rightAlt|shift)).Rets(K('ẞ')),

		Args(charKeyEvent(uint16(r1), 0)).Rets(surrogateKeyEvent{r1}),
		Args(charKeyEvent(uint16(r2), 0)).Rets(surrogateKeyEvent{r2}),

		Args(funcKeyEvent(0x1b, 0)).Rets(K('[', ui.Ctrl)),

		// Functional key with modifiers
		Args(funcKeyEvent(0x08, 0)).Rets(K(ui.Backspace)),
		Args(funcKeyEvent(0x08, leftCtrl)).Rets(K(ui.Backspace, ui.Ctrl)),
		Args(funcKeyEvent(0x08, leftCtrl|leftAlt|shift)).Rets(K(ui.Backspace, ui.Ctrl, ui.Alt, ui.Shift)),

		// Functional keys with an alphanumeric base
		Args(funcKeyEvent('2', leftCtrl)).Rets(K('2', ui.Ctrl)),
		Args(funcKeyEvent('A', leftCtrl)).Rets(K('A', ui.Ctrl)),
		Args(funcKeyEvent('A', leftAlt)).Rets(K('a', ui.Alt)),

		// Unrecognized functional key
		Args(funcKeyEvent(0, 0)).Rets(nil),
	)
}

func charKeyEvent(r uint16, mod uint32) *ewindows.KeyEvent {
	return &ewindows.KeyEvent{
		BKeyDown: 1, DwControlKeyState: mod, UChar: [2]byte{byte(r), byte(r >> 8)}}
}

func funcKeyEvent(code uint16, mod uint32) *ewindows.KeyEvent {
	return &ewindows.KeyEvent{
		BKeyDown: 1, DwControlKeyState: mod, WVirtualKeyCode: code}
}

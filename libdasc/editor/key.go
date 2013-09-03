package editor

type Key struct {
	rune
	Ctrl bool
	Alt bool
}

const (
	F1 rune = -1-iota
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12

	Escape // ^[
	Backspace // ^?

	Up // ^[OA
	Down // ^[OB
	Right // ^[OC
	Left // ^[OD

	Home // ^[[1~
	Insert // ^[[2~
	Delete // ^[[3~
	End // ^[[4~
	PageUp // ^[[5~
	PageDown // ^[[6~
)

func decodeKey(s string) (k Key, pending bool, err error) {
	for _, r := range s {
		/*
		if r == 0x1b {
			// ^[, or Escape
			if i == len(s) - 1 {
				k = Key{rune: '[': Ctrl: true}
				pending = true
				return
			} else {
			}
		} else */ if r <= 0x1f {
			k = Key{rune: r+0x40, Ctrl: true}
			return
		} else {
			k = Key{rune: r}
			return
		}
	}
	return
}

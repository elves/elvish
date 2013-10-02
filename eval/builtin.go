package eval

type ioType byte

const (
	fileIO ioType = iota // Default IO type. Corresponds to io.f.
	chanIO // Corresponds to io.ch.
	unusedIO
)

type builtinFunc func([]string, [3]*io)

type builtin struct {
	f builtinFunc
	ioTypes [3]ioType
}

var builtins = map[string]builtin {
	"put": builtin{implPut, [3]ioType{unusedIO, chanIO}},
}

func implPut(args []string, ios [3]*io) {
	out := ios[1].ch
	for i := 1; i < len(args); i++ {
		out <- args[i]
	}
}

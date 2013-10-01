package eval

type ioType byte

const (
	// Default IO type. Corresponding io argument is a uintptr representing a
	// Unix fd.
	fdIO ioType = iota
	chanIO // Corresponding io argument is a chan string (for now).
	unusedIO // Corresponding io argument is not used at all.
)

type builtinFunc func([]string, [3]interface{})

type builtin struct {
	f builtinFunc
	ioTypes [3]ioType
}

var builtins = map[string]builtin {
	"put": builtin{implPut, [3]ioType{unusedIO, chanIO}},
}

func implPut(args []string, ios [3]interface{}) {
	out := ios[1].(chan string)
	for i := 1; i < len(args); i++ {
		out <- args[i]
	}
}

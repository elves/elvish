package eval

type commandType int

const (
	commandBuiltinFunction commandType = iota
	commandBuiltinSpecial
	commandDefinedFunction
	commandExternal
)

type formAnnotation struct {
	streamTypes       [2]StreamType
	commandType       commandType
	builtinFunc       *builtinFunc
	builtinSpecial    *builtinSpecial
	specialAnnotation interface{}
}

type pipelineAnnotation struct {
	bounds    [2]StreamType
	internals []StreamType
}

type closureAnnotation struct {
	enclosed map[string]Type
	bounds   [2]StreamType
}

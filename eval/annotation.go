package eval

type commandType int

const (
	commandBuiltinFunction commandType = iota
	commandBuiltinSpecial
	commandDefinedFunction
	commandClosure
	commandExternal
)

type formAnnotation struct {
	streamTypes    [2]StreamType
	commandType    commandType
	builtinFunc    *builtinFunc
	builtinSpecial *builtinSpecial
	specialOp      strOp
}

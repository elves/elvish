package parse

type ContextType int

const (
	NewFactorContext ContextType = iota
	FilenameContext
	CommandContext
	VariableNameContext
)

type Context struct {
	Typ    ContextType
	Prefix string
}

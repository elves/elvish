package parse

type ContextType int

const (
	NewTermContext ContextType = iota
	FilenameContext
	CommandContext
	VariableNameContext
)

type Context struct {
	Typ    ContextType
	Prefix string
}

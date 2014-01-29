package parse

type ContextType int

const (
	UnknownContext ContextType = iota
	FilenameContext
	CommandContext
	VariableNameContext
	NewFactorContext // Have to begin a new factor
)

type Context struct {
	Typ    ContextType
	Prefix string
}

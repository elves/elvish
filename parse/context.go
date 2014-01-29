package parse

type ContextType int

const (
	UnknownContext ContextType = iota
	FilenameContext
	CommandContext
)

type Context struct {
	Typ ContextType
	Prefix string
}

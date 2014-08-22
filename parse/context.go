package parse

// ContextType categorizes Context.
type ContextType int

// ContextType values.
const (
	UnknownContext ContextType = iota
	CommandContext
	ArgContext
	NewArgContext
	TableKeyContext
	TableValueContext
	TableElemContext
	RedirFilenameContext
	StatusRedirContext
)

// Context contains an incomplete FormNode and indication of what to complete.
type Context struct {
	Typ  ContextType
	Form *FormNode
}

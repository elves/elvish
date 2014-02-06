package parse

// ContextType categorizes Context.
type ContextType int

// ContextType values.
const (
	CommandContext ContextType = iota
	ArgContext
	RedirFilenameContext
)

// Context contains information from the AST useful for tab completion.
type Context struct {
	Typ         ContextType
	CommandTerm *ListNode
	PrevTerms   *ListNode
	PrevFactors *ListNode
	ThisFactor  *FactorNode
}

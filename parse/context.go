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

// Isomorph compares two Contexts, ignoring all Pos'es.
func (c *Context) isomorph(c2 *Context) bool {
	return c.Typ == c2.Typ && c.CommandTerm.isomorph(c2.CommandTerm) && c.PrevTerms.isomorph(c2.PrevTerms) && c.PrevFactors.isomorph(c2.PrevFactors) && c.ThisFactor.isomorph(c2.ThisFactor)
}

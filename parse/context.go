package parse

// Context contains information from the AST useful for tab completion.
type Context struct {
	PrevTerms   *ListNode
	PrevFactors *ListNode
	ThisFactor  *FactorNode
}

// Isomorph compares two Contexts, ignoring all Pos'es.
func (c *Context) isomorph(c2 *Context) bool {
	return c.PrevTerms.isomorph(c2.PrevTerms) && c.PrevFactors.isomorph(c2.PrevFactors) && c.ThisFactor.isomorph(c2.ThisFactor)
}

package parse

// Context contains information from the AST useful for tab completion.
type Context struct {
	PrevTerms   *ListNode
	PrevFactors *ListNode
	ThisFactor  *FactorNode
}

// Isomorph compares two Contexts, ignoring all Pos'es.
func (c *Context) Isomorph(c2 *Context) bool {
	return c.PrevTerms.Isomorph(c2.PrevTerms) && c.PrevFactors.Isomorph(c2.PrevFactors) && c.ThisFactor.Isomorph(c2.ThisFactor)
}

package parse

type Context struct {
	PrevTerms   *ListNode
	PrevFactors *ListNode
	ThisFactor  *FactorNode
}

func (c *Context) Isomorph(c2 *Context) bool {
	return c.PrevTerms.Isomorph(c2.PrevTerms) && c.PrevFactors.Isomorph(c2.PrevFactors) && c.ThisFactor.Isomorph(c2.ThisFactor)
}

package parse

type Context interface {
	ToComplete() string
}

type context struct {
	toComplete string
}

func (ctx *context) ToComplete() string {
	return ctx.toComplete
}

type ArgContext struct {
	context
}

func NewArgContext(tc string) *ArgContext {
	return &ArgContext{context{tc}}
}

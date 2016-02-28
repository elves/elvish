package edit

// ArgCompleter is an argument completer. Its Complete method is called with
// the head of the form, a list of preceding arguments and the (possibly empty)
// current argument. It should return a list of candidates for the current
// argument and errors.
type ArgCompleter interface {
	Complete(*ArgContext) ([]*candidate, error)
}

type ArgContext struct {
	head    string
	args    []string
	current string
	ed      *Editor
}

type FuncArgCompleter struct {
	impl func(*ArgContext) ([]*candidate, error)
}

func (fac FuncArgCompleter) Complete(ctx *ArgContext) ([]*candidate, error) {
	return fac.impl(ctx)
}

var DefaultArgCompleter = ""
var argCompleter map[string]ArgCompleter

func init() {
	argCompleter = map[string]ArgCompleter{
		DefaultArgCompleter: FuncArgCompleter{complFilename},
		"sudo":              FuncArgCompleter{complSudo},
	}
}

func completeArg(ctx *ArgContext) ([]*candidate, error) {
	Logger.Printf("completing argument: %q %q %q", ctx.head, ctx.args, ctx.current)
	compl, ok := argCompleter[ctx.head]
	if !ok {
		compl = argCompleter[DefaultArgCompleter]
	}
	return compl.Complete(ctx)
}

func complFilename(ctx *ArgContext) ([]*candidate, error) {
	return complFilenameInner(ctx.current, false)
}

func complSudo(ctx *ArgContext) ([]*candidate, error) {
	if len(ctx.args) == 0 {
		return complFormHeadInner(ctx.current, ctx.ed)
	}
	return completeArg(&ArgContext{ctx.args[0], ctx.args[1:], ctx.current, ctx.ed})
}

package edit

// ArgCompleter is an argument completer. Its Complete method is called with all
// words of the form. There are at least two words: the first one being the form
// head and the last word being the current argument to complete. It should
// return a list of candidates for the current argument and errors.
type ArgCompleter interface {
	Complete([]string, *Editor) ([]*candidate, error)
}

type FuncArgCompleter struct {
	impl func([]string, *Editor) ([]*candidate, error)
}

func (fac FuncArgCompleter) Complete(words []string, ed *Editor) ([]*candidate, error) {
	return fac.impl(words, ed)
}

// CompleterTable provides $le:completer. It implements eval.IndexSetter.
type CompleterTable map[string]ArgCompleter

func (CompleterTable) Kind() string {
	return "map"
}

func (ct CompleterTable) Repr(indent int) string {
	return ""
}

var DefaultArgCompleter = ""
var argCompleter map[string]ArgCompleter

func init() {
	argCompleter = map[string]ArgCompleter{
		DefaultArgCompleter: FuncArgCompleter{complFilename},
		"sudo":              FuncArgCompleter{complSudo},
	}
}

func completeArg(words []string, ed *Editor) ([]*candidate, error) {
	Logger.Printf("completing argument: %q", words)
	compl, ok := argCompleter[words[0]]
	if !ok {
		compl = argCompleter[DefaultArgCompleter]
	}
	return compl.Complete(words, ed)
}

func complFilename(words []string, ed *Editor) ([]*candidate, error) {
	return complFilenameInner(words[len(words)-1], false)
}

func complSudo(words []string, ed *Editor) ([]*candidate, error) {
	if len(words) == 2 {
		return complFormHeadInner(words[1], ed)
	}
	return completeArg(words[1:], ed)
}

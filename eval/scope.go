package eval

// Scope represents a lexical scope. It holds variables and imports.
type Scope struct {
	Names Namespace
	Uses  map[string]Namespace
}

func makeScope() Scope {
	return Scope{Namespace{}, map[string]Namespace{}}
}

// Namespace is a map from names to variables.
type Namespace map[string]Variable

// staticScope represents static information of a staticScope.
// TODO(xiaq): Represent Scope.Uses as well.
type staticScope map[string]bool

func makeStaticScope(s Namespace) staticScope {
	sc := staticScope{}
	for name := range s {
		sc[name] = true
	}
	return sc
}

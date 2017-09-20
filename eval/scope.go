package eval

// Scope represents a lexical scope. It holds variables and imports.
type Scope struct {
	Names Namespace
	Uses  map[string]Namespace
}

// Namespace is a map from names to variables.
type Namespace map[string]Variable

func makeScope() Scope {
	return Scope{Namespace{}, map[string]Namespace{}}
}

// staticScope represents static information of a staticScope.
// TODO(xiaq): Represent Scope.Uses as well.
type staticScope struct {
	Names map[string]bool
	Uses  map[string]bool
}

func makeStaticScope() staticScope {
	return staticScope{map[string]bool{}, map[string]bool{}}
}

func (s Scope) static() staticScope {
	ss := makeStaticScope()
	for name := range s.Names {
		ss.Names[name] = true
	}
	for name := range s.Uses {
		ss.Uses[name] = true
	}
	return ss
}

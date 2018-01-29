package eval

func resolve(s string, ec *Frame) Fn {
	// NOTE: This needs to be kept in sync with the resolution algorithm used in
	// (*compiler).form.

	// Try variable
	explode, ns, name := ParseVariableRef(s)
	if !explode {
		if v := ec.ResolveVar(ns, name+FnSuffix); v != nil {
			if caller, ok := v.Get().(Fn); ok {
				return caller
			}
		}
	}

	// External command
	return ExternalCmd{s}
}

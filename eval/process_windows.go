package eval

// Process control functions in Windows. These are all NOPs.
func ignoreTTOU()        {}
func unignoreTTOU()      {}
func putSelfInFg() error { return nil }

package parse

// IsChunk reports whether the node has type *Chunk.
func IsChunk(n Node) bool {
	_, ok := n.(*Chunk)
	return ok
}

// GetChunk returns the node cast to *Chunk if the node has that type, or nil otherwise.
func GetChunk(n Node) *Chunk {
	if nn, ok := n.(*Chunk); ok {
		return nn
	}
	return nil
}

// IsPipeline reports whether the node has type *Pipeline.
func IsPipeline(n Node) bool {
	_, ok := n.(*Pipeline)
	return ok
}

// GetPipeline returns the node cast to *Pipeline if the node has that type, or nil otherwise.
func GetPipeline(n Node) *Pipeline {
	if nn, ok := n.(*Pipeline); ok {
		return nn
	}
	return nil
}

// IsForm reports whether the node has type *Form.
func IsForm(n Node) bool {
	_, ok := n.(*Form)
	return ok
}

// GetForm returns the node cast to *Form if the node has that type, or nil otherwise.
func GetForm(n Node) *Form {
	if nn, ok := n.(*Form); ok {
		return nn
	}
	return nil
}

// IsAssignment reports whether the node has type *Assignment.
func IsAssignment(n Node) bool {
	_, ok := n.(*Assignment)
	return ok
}

// GetAssignment returns the node cast to *Assignment if the node has that type, or nil otherwise.
func GetAssignment(n Node) *Assignment {
	if nn, ok := n.(*Assignment); ok {
		return nn
	}
	return nil
}

// IsExitusRedir reports whether the node has type *ExitusRedir.
func IsExitusRedir(n Node) bool {
	_, ok := n.(*ExitusRedir)
	return ok
}

// GetExitusRedir returns the node cast to *ExitusRedir if the node has that type, or nil otherwise.
func GetExitusRedir(n Node) *ExitusRedir {
	if nn, ok := n.(*ExitusRedir); ok {
		return nn
	}
	return nil
}

// IsRedir reports whether the node has type *Redir.
func IsRedir(n Node) bool {
	_, ok := n.(*Redir)
	return ok
}

// GetRedir returns the node cast to *Redir if the node has that type, or nil otherwise.
func GetRedir(n Node) *Redir {
	if nn, ok := n.(*Redir); ok {
		return nn
	}
	return nil
}

// IsCompound reports whether the node has type *Compound.
func IsCompound(n Node) bool {
	_, ok := n.(*Compound)
	return ok
}

// GetCompound returns the node cast to *Compound if the node has that type, or nil otherwise.
func GetCompound(n Node) *Compound {
	if nn, ok := n.(*Compound); ok {
		return nn
	}
	return nil
}

// IsIndexing reports whether the node has type *Indexing.
func IsIndexing(n Node) bool {
	_, ok := n.(*Indexing)
	return ok
}

// GetIndexing returns the node cast to *Indexing if the node has that type, or nil otherwise.
func GetIndexing(n Node) *Indexing {
	if nn, ok := n.(*Indexing); ok {
		return nn
	}
	return nil
}

// IsArray reports whether the node has type *Array.
func IsArray(n Node) bool {
	_, ok := n.(*Array)
	return ok
}

// GetArray returns the node cast to *Array if the node has that type, or nil otherwise.
func GetArray(n Node) *Array {
	if nn, ok := n.(*Array); ok {
		return nn
	}
	return nil
}

// IsPrimary reports whether the node has type *Primary.
func IsPrimary(n Node) bool {
	_, ok := n.(*Primary)
	return ok
}

// GetPrimary returns the node cast to *Primary if the node has that type, or nil otherwise.
func GetPrimary(n Node) *Primary {
	if nn, ok := n.(*Primary); ok {
		return nn
	}
	return nil
}

// IsMapPair reports whether the node has type *MapPair.
func IsMapPair(n Node) bool {
	_, ok := n.(*MapPair)
	return ok
}

// GetMapPair returns the node cast to *MapPair if the node has that type, or nil otherwise.
func GetMapPair(n Node) *MapPair {
	if nn, ok := n.(*MapPair); ok {
		return nn
	}
	return nil
}

// IsSep reports whether the node has type *Sep.
func IsSep(n Node) bool {
	_, ok := n.(*Sep)
	return ok
}

// GetSep returns the node cast to *Sep if the node has that type, or nil otherwise.
func GetSep(n Node) *Sep {
	if nn, ok := n.(*Sep); ok {
		return nn
	}
	return nil
}

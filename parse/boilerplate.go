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

func (n *Chunk) addToPipelines(ch *Pipeline) {
	n.Pipelines = append(n.Pipelines, ch)
	addChild(n, ch)
}

// ParseChunk parses a node of type *Chunk.
func ParseChunk(ps *Parser) *Chunk {
	n := &Chunk{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *Pipeline) addToForms(ch *Form) {
	n.Forms = append(n.Forms, ch)
	addChild(n, ch)
}

// ParsePipeline parses a node of type *Pipeline.
func ParsePipeline(ps *Parser) *Pipeline {
	n := &Pipeline{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *Form) addToAssignments(ch *Assignment) {
	n.Assignments = append(n.Assignments, ch)
	addChild(n, ch)
}

func (n *Form) setHead(ch *Compound) {
	n.Head = ch
	addChild(n, ch)
}

func (n *Form) addToVars(ch *Compound) {
	n.Vars = append(n.Vars, ch)
	addChild(n, ch)
}

func (n *Form) addToArgs(ch *Compound) {
	n.Args = append(n.Args, ch)
	addChild(n, ch)
}

func (n *Form) addToOpts(ch *MapPair) {
	n.Opts = append(n.Opts, ch)
	addChild(n, ch)
}

func (n *Form) addToRedirs(ch *Redir) {
	n.Redirs = append(n.Redirs, ch)
	addChild(n, ch)
}

func (n *Form) setExitusRedir(ch *ExitusRedir) {
	n.ExitusRedir = ch
	addChild(n, ch)
}

// ParseForm parses a node of type *Form.
func ParseForm(ps *Parser) *Form {
	n := &Form{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *Assignment) setLeft(ch *Indexing) {
	n.Left = ch
	addChild(n, ch)
}

func (n *Assignment) setRight(ch *Compound) {
	n.Right = ch
	addChild(n, ch)
}

// ParseAssignment parses a node of type *Assignment.
func ParseAssignment(ps *Parser) *Assignment {
	n := &Assignment{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *ExitusRedir) setDest(ch *Compound) {
	n.Dest = ch
	addChild(n, ch)
}

// ParseExitusRedir parses a node of type *ExitusRedir.
func ParseExitusRedir(ps *Parser) *ExitusRedir {
	n := &ExitusRedir{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *Redir) setLeft(ch *Compound) {
	n.Left = ch
	addChild(n, ch)
}

func (n *Redir) setRight(ch *Compound) {
	n.Right = ch
	addChild(n, ch)
}

// ParseRedir parses a node of type *Redir.
func ParseRedir(ps *Parser, dest *Compound) *Redir {
	n := &Redir{node: node{begin: ps.pos}}
	n.parse(ps, dest)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *Compound) addToIndexings(ch *Indexing) {
	n.Indexings = append(n.Indexings, ch)
	addChild(n, ch)
}

// ParseCompound parses a node of type *Compound.
func ParseCompound(ps *Parser, ctx ExprCtx) *Compound {
	n := &Compound{node: node{begin: ps.pos}}
	n.parse(ps, ctx)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *Indexing) setHead(ch *Primary) {
	n.Head = ch
	addChild(n, ch)
}

func (n *Indexing) addToIndicies(ch *Array) {
	n.Indicies = append(n.Indicies, ch)
	addChild(n, ch)
}

// ParseIndexing parses a node of type *Indexing.
func ParseIndexing(ps *Parser, ctx ExprCtx) *Indexing {
	n := &Indexing{node: node{begin: ps.pos}}
	n.parse(ps, ctx)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *Array) addToCompounds(ch *Compound) {
	n.Compounds = append(n.Compounds, ch)
	addChild(n, ch)
}

// ParseArray parses a node of type *Array.
func ParseArray(ps *Parser, allowSemicolon bool) *Array {
	n := &Array{node: node{begin: ps.pos}}
	n.parse(ps, allowSemicolon)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *Primary) addToElements(ch *Compound) {
	n.Elements = append(n.Elements, ch)
	addChild(n, ch)
}

func (n *Primary) setChunk(ch *Chunk) {
	n.Chunk = ch
	addChild(n, ch)
}

func (n *Primary) addToMapPairs(ch *MapPair) {
	n.MapPairs = append(n.MapPairs, ch)
	addChild(n, ch)
}

func (n *Primary) addToBraced(ch *Compound) {
	n.Braced = append(n.Braced, ch)
	addChild(n, ch)
}

// ParsePrimary parses a node of type *Primary.
func ParsePrimary(ps *Parser, ctx ExprCtx) *Primary {
	n := &Primary{node: node{begin: ps.pos}}
	n.parse(ps, ctx)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

func (n *MapPair) setKey(ch *Compound) {
	n.Key = ch
	addChild(n, ch)
}

func (n *MapPair) setValue(ch *Compound) {
	n.Value = ch
	addChild(n, ch)
}

// ParseMapPair parses a node of type *MapPair.
func ParseMapPair(ps *Parser) *MapPair {
	n := &MapPair{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
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

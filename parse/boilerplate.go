package parse

func IsChunk(n Node) bool {
	_, ok := n.(*Chunk)
	return ok
}

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

func ParseChunk(ps *Parser) *Chunk {
	n := &Chunk{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsPipeline(n Node) bool {
	_, ok := n.(*Pipeline)
	return ok
}

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

func ParsePipeline(ps *Parser) *Pipeline {
	n := &Pipeline{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsForm(n Node) bool {
	_, ok := n.(*Form)
	return ok
}

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

func ParseForm(ps *Parser) *Form {
	n := &Form{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsAssignment(n Node) bool {
	_, ok := n.(*Assignment)
	return ok
}

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

func ParseAssignment(ps *Parser) *Assignment {
	n := &Assignment{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsExitusRedir(n Node) bool {
	_, ok := n.(*ExitusRedir)
	return ok
}

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

func ParseExitusRedir(ps *Parser) *ExitusRedir {
	n := &ExitusRedir{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsRedir(n Node) bool {
	_, ok := n.(*Redir)
	return ok
}

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

func ParseRedir(ps *Parser, dest *Compound) *Redir {
	n := &Redir{node: node{begin: ps.pos}}
	n.parse(ps, dest)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsCompound(n Node) bool {
	_, ok := n.(*Compound)
	return ok
}

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

func ParseCompound(ps *Parser, ctx ExprCtx) *Compound {
	n := &Compound{node: node{begin: ps.pos}}
	n.parse(ps, ctx)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsIndexing(n Node) bool {
	_, ok := n.(*Indexing)
	return ok
}

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

func ParseIndexing(ps *Parser, ctx ExprCtx) *Indexing {
	n := &Indexing{node: node{begin: ps.pos}}
	n.parse(ps, ctx)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsArray(n Node) bool {
	_, ok := n.(*Array)
	return ok
}

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

func ParseArray(ps *Parser, allowSemicolon bool) *Array {
	n := &Array{node: node{begin: ps.pos}}
	n.parse(ps, allowSemicolon)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsPrimary(n Node) bool {
	_, ok := n.(*Primary)
	return ok
}

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

func ParsePrimary(ps *Parser, ctx ExprCtx) *Primary {
	n := &Primary{node: node{begin: ps.pos}}
	n.parse(ps, ctx)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsMapPair(n Node) bool {
	_, ok := n.(*MapPair)
	return ok
}

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

func ParseMapPair(ps *Parser) *MapPair {
	n := &MapPair{node: node{begin: ps.pos}}
	n.parse(ps)
	n.end = ps.pos
	n.sourceText = ps.src[n.begin:n.end]
	return n
}

func IsSep(n Node) bool {
	_, ok := n.(*Sep)
	return ok
}

func GetSep(n Node) *Sep {
	if nn, ok := n.(*Sep); ok {
		return nn
	}
	return nil
}

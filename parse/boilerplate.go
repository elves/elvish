package parse

func (n *MapPair) setKey(ch *Compound) {
	n.Key = ch
	addChild(n, ch)
}

func (n *MapPair) setValue(ch *Compound) {
	n.Value = ch
	addChild(n, ch)
}

func parseMapPair(rd *reader, cut runePred) *MapPair {
	n := &MapPair{node: node{begin: rd.pos}}
	n.parse(rd, cut)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *Primary) setList(ch *Array) {
	n.List = ch
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

func parsePrimary(rd *reader, cut runePred) *Primary {
	n := &Primary{node: node{begin: rd.pos}}
	n.parse(rd, cut)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *Indexed) setHead(ch *Primary) {
	n.Head = ch
	addChild(n, ch)
}

func (n *Indexed) addToIndicies(ch *Array) {
	n.Indicies = append(n.Indicies, ch)
	addChild(n, ch)
}

func parseIndexed(rd *reader, cut runePred) *Indexed {
	n := &Indexed{node: node{begin: rd.pos}}
	n.parse(rd, cut)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *Compound) addToIndexeds(ch *Indexed) {
	n.Indexeds = append(n.Indexeds, ch)
	addChild(n, ch)
}

func parseCompound(rd *reader, cut runePred) *Compound {
	n := &Compound{node: node{begin: rd.pos}}
	n.parse(rd, cut)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *Array) addToCompounds(ch *Compound) {
	n.Compounds = append(n.Compounds, ch)
	addChild(n, ch)
}

func parseArray(rd *reader) *Array {
	n := &Array{node: node{begin: rd.pos}}
	n.parse(rd)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *Redir) setDest(ch *Compound) {
	n.Dest = ch
	addChild(n, ch)
}

func (n *Redir) setSource(ch *Compound) {
	n.Source = ch
	addChild(n, ch)
}

func parseRedir(rd *reader, cut runePred, dest *Compound) *Redir {
	n := &Redir{node: node{begin: rd.pos}}
	n.parse(rd, cut, dest)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *ExitusRedir) setDest(ch *Compound) {
	n.Dest = ch
	addChild(n, ch)
}

func parseExitusRedir(rd *reader, cut runePred) *ExitusRedir {
	n := &ExitusRedir{node: node{begin: rd.pos}}
	n.parse(rd, cut)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *Form) addToAssignments(ch *Assignment) {
	n.Assignments = append(n.Assignments, ch)
	addChild(n, ch)
}

func (n *Form) setHead(ch *Compound) {
	n.Head = ch
	addChild(n, ch)
}

func (n *Form) addToArgs(ch *Compound) {
	n.Args = append(n.Args, ch)
	addChild(n, ch)
}

func (n *Form) addToNamedArgs(ch *MapPair) {
	n.NamedArgs = append(n.NamedArgs, ch)
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

func parseForm(rd *reader, cut runePred) *Form {
	n := &Form{node: node{begin: rd.pos}}
	n.parse(rd, cut)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *Assignment) setDst(ch *Indexed) {
	n.Dst = ch
	addChild(n, ch)
}

func (n *Assignment) setSrc(ch *Compound) {
	n.Src = ch
	addChild(n, ch)
}

func parseAssignment(rd *reader, cut runePred) *Assignment {
	n := &Assignment{node: node{begin: rd.pos}}
	n.parse(rd, cut)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *Pipeline) addToForms(ch *Form) {
	n.Forms = append(n.Forms, ch)
	addChild(n, ch)
}

func parsePipeline(rd *reader, cut runePred) *Pipeline {
	n := &Pipeline{node: node{begin: rd.pos}}
	n.parse(rd, cut)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

func (n *Chunk) addToPipelines(ch *Pipeline) {
	n.Pipelines = append(n.Pipelines, ch)
	addChild(n, ch)
}

func parseChunk(rd *reader, cut runePred) *Chunk {
	n := &Chunk{node: node{begin: rd.pos}}
	n.parse(rd, cut)
	n.end = rd.pos
	n.sourceText = rd.src[n.begin:n.end]
	return n
}

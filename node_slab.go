package incr

// nodeSlab hands out [Node] values from contiguous chunks.
//
// A node would otherwise cost two allocations: the concrete node type, and the metadata it
// points at. The metadata is the larger of the two and every node needs one, so carving them
// from chunks turns most of those allocations into a bounds check and an index bump -- a
// bind rebuilding an eight leaf subgraph went from 73 allocations and 11.7KB to 34 and
// 1.8KB. It also puts nodes created together next to each other in memory, which is what a
// stabilization walks, and measured that is worth more than the cheaper allocation: a wide
// graph's construction and updating every input each improve about 10%, and a bind swapping
// its subgraph about 20%.
//
// Chunks are ordinary heap objects, so this changes the granularity of reclamation rather
// than escaping it. A chunk is collected once every node in it is unreachable, which means
// one live node keeps its whole chunk alive; that is the cost of the arrangement, and it is
// why chunks start small.
//
// A graph's own slab only ever moves forward, so a slot it hands out is never reused. A bind
// scope's slab is reset on each rebuild and does reissue slots; see reset for the contract
// that relies on.
type nodeSlab struct {
	// chunks holds every chunk handed out since the last reset. A generation that is rebuilt
	// repeatedly reuses these rather than allocating new ones, so in the steady state a
	// bind's right-hand side costs no allocation at all for its metadata. A bind alternates
	// two slabs, so reuse begins on its third rebuild.
	chunks [][]Node
	// chunk is the index into chunks currently being filled, and next the bump position
	// within it.
	chunk int
	next  int
	// size is the length of the next chunk to allocate. It starts small and doubles, so a
	// bind whose right-hand side is two nodes does not reserve room for hundreds, while a
	// large graph still amortizes down to one allocation per chunk.
	size int
}

const (
	// nodeSlabMinChunk is deliberately tiny. Every bind scope has its own slab, so a graph
	// with thousands of binds whose right-hand sides are one or two nodes each pays this
	// minimum thousands of times over: at 8 it cost 27% on such a graph, at 2 it costs
	// nothing. Larger scopes reach a useful chunk size within a few doublings anyway.
	nodeSlabMinChunk = 2
	// nodeSlabMaxChunk caps how much a single surviving node can pin. At 384 bytes per node
	// a 256 slot chunk is about 96KB.
	nodeSlabMaxChunk = 256
)

// get returns the next free slot, allocating a chunk when the current ones are spent.
//
// The slot is scrubbed before it is handed out, because after a reset a slot may be issued
// a second time and newNodeIn only writes the fields that are not zero valued.
func (s *nodeSlab) get() *Node {
	if s.chunk < len(s.chunks) {
		current := s.chunks[s.chunk]
		if s.next == len(current) {
			s.chunk++
			s.next = 0
			return s.get()
		}
		n := &current[s.next]
		s.next++
		*n = Node{}
		return n
	}
	if s.size == 0 {
		s.size = nodeSlabMinChunk
	}
	s.chunks = append(s.chunks, make([]Node, s.size))
	s.chunk = len(s.chunks) - 1
	s.next = 0
	if s.size < nodeSlabMaxChunk {
		s.size *= 2
	}
	return s.get()
}

// reset makes every slot handed out since the last reset available again, reusing the same
// memory rather than releasing it.
//
// This is what a bind does when it rebuilds its right-hand side. Reusing the memory is the
// point: the generation being replaced is dead by then, and reissuing its slots means a
// steady-state rebuild allocates nothing and works in memory that is already warm. Dropping
// the chunks instead would allocate a fresh generation every rebuild and hand the old one to
// the collector, which measured 1.65x slower than not doing this at all.
//
// The cost is that a slot can be issued twice, so a caller retaining an [Incr] created
// inside a bind's scope past a rebuild would observe another node's metadata. That is
// already the documented contract for bind scopes: nodes created there are released when the
// right-hand side is rewritten.
func (s *nodeSlab) reset() {
	s.chunk = 0
	s.next = 0
}

// newNodeIn initializes a slot from the given slab as a node of the given kind.
//
// This is the internal counterpart to [NewNode], which allocates on its own and is kept
// that way for node types defined outside this package.
func newNodeIn(slab *nodeSlab, kind string) *Node {
	n := slab.get()
	n.kind = kind
	n.valid = true // start out valid!
	n.height = HeightUnset
	n.heightInRecomputeHeap = HeightUnset
	n.heightInAdjustHeightsHeap = HeightUnset
	n.graphIndex = -1
	return n
}

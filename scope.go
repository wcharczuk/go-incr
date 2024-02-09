package incr

// WithinBindScope updates a node's createdIn scope to reflect a new
// inner-most bind scope applied by a bind.
//
// It operates off go contexts to make the threading of the scopes
// as transparent as possible, but this helper is exported in special
// cases where a node might have been already created but you still
// want to add it to a bind scope (e.g. with node caches).
//
// The scope itself is created by a bind at its construction, and will
// hold a reference to both the bind node itself and the bind's
// left-hand-side or input node. All nodes created within the bind's function
// will be associated with its scope unless they were created _outside_
// that scope and are simply referenced by nodes created by the bind.
func WithinScope[A INode](scope Scope, node A) A {
	scope.AddNode(node)
	return node
}

// Scope is a context that can add nodes.
type Scope interface {
	IsTop() bool
	IsNecessary() bool
	IsValid() bool
	Height() int
	AddNode(INode)
	Nodes() []INode
	Graph() *Graph
}

var (
	_ Scope = (*bindScope)(nil)
)

type bindScope struct {
	bind     IBind
	rhsNodes []INode
}

func (bs *bindScope) IsTop() bool {
	return false
}

func (bs *bindScope) IsNecessary() bool {
	return bs.bind.Node().isNecessary()
}

func (bs *bindScope) IsValid() bool {
	return bs.bind.Node().isValid
}

func (bs *bindScope) Height() int {
	return bs.bind.BindChange().Node().height
}

func (bs *bindScope) AddNode(node INode) {
	bs.rhsNodes = append(bs.rhsNodes, node)
}

func (bs *bindScope) Nodes() []INode {
	return bs.rhsNodes
}

func (bs *bindScope) Graph() *Graph {
	return bs.bind.Node().graph
}

package incr

import "fmt"

// Root is is the top level node bind scope.
//
// When in doubt, pass this as the scope argument
// to a node constructor, e.g.
//
//	myVar := incr.Var(incr.Root(), "hello world!")
//
// Do _not_ pass this as the scope argument to node
// constructors within Bind functions!
func Root() *BindScope {
	return _root
}

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
func WithinBindScope[A INode](scope *BindScope, node A) A {
	addNodeToBindScope(scope, node)
	return node
}

// BindScope is the scope that nodes are created in.
//
// Its either nil or the most recent bind.
type BindScope struct {
	root     bool
	bind     INode
	lhsNodes []INode
	rhsNodes []INode
}

// Height returns the maximum height of any
// of the lhs nodes.
func (bs *BindScope) Height() (out int) {
	for _, lhs := range bs.lhsNodes {
		if out < lhs.Node().height {
			out = lhs.Node().height
		}
	}
	return
}

func (bs *BindScope) String() string {
	if bs == nil || bs.root {
		return "{root}"
	}
	return fmt.Sprintf("{%v}", bs.bind)
}

var _root = &BindScope{root: true}

func addNodeToBindScope(scope *BindScope, node INode) {
	if scope == nil || scope.root {
		return
	}
	if node.Node().createdIn != nil {
		node.Node().createdIn.rhsNodes = remove(node.Node().createdIn.rhsNodes, node.Node().id)
	}
	node.Node().createdIn = scope
	scope.rhsNodes = append(scope.rhsNodes, node)
}

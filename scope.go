package incr

import "fmt"

// WithinScope updates a node's createdIn scope to reflect a new inner-most
// bind scope applied by a bind.
//
// The scope itself is created by a bind at its construction, and will
// hold a reference to both the bind node itself and the bind's
// left-hand-side or input node. All nodes created within the bind's function
// will be associated with its scope unless they were created outside
// that scope and are simply referenced by nodes created by the bind.
//
// You shouldn't need to call [WithinScope] in practice typically, as it is
// sufficient to pass the [Bind] function's provided scope to node constructors
// to associate the scopes correctly. This method is exported for advanced use
// cases where you want to manage scopes manually.
func WithinScope[A INode](scope Scope, node A) A {
	node.Node().createdIn = scope
	if scope != nil && scope.isTopScope() {
		return node
	}
	scope.addScopeNode(node)
	return node
}

// GraphForScope returns the graph for a given scope.
//
// It is used for advanced situations where you need to refer back to the underlying graph
// of a scope but may not have a reference available (e.g. in generalized bind functions).
func GraphForScope(s Scope) *Graph {
	return s.scopeGraph()
}

// GraphForNode returns the graph for a given node as derrived through
// the scope it was created in, which must return a graph reference.
func GraphForNode(node INode) *Graph {
	if node == nil {
		return nil
	}
	return node.Node().createdIn.scopeGraph()
}

// Scope is a type that's used to track which nodes are created by which "areas" of the graph.
//
// When in doubt, if you see a scope argument you should pass the [Graph] itself.
//
// If you're within a bind, you should pass the scope that is passed to your bind function.
type Scope interface {
	isTopScope() bool
	isScopeValid() bool
	isScopeNecessary() bool
	scopeGraph() *Graph
	scopeHeight() int
	addScopeNode(INode)
	fmt.Stringer
}

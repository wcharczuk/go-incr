package incr

import "fmt"

// WithinScope updates a node's createdIn scope to reflect a new
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
	maybeAddScopeNode(scope, node)
	return node
}

// Scope is a type that's used to track which nodes are created by which "areas" of the graph.
//
// When in doubt, if you see a scope argument you should pass the `Graph` itself.
//
// If you're within a bind, you should pass the scope that is passed to your bind function.
type Scope interface {
	isTopScope() bool
	isScopeValid() bool
	isScopeNecessary() bool
	scopeGraph() *Graph
	scopeHeight() int
	addScopeNode(INode)
	removeScopeNode(INode)
	fmt.Stringer
}

func graphFromScope(node INode) *Graph {
	if node == nil {
		return nil
	}
	return node.Node().createdIn.scopeGraph()
}

func maybeRemoveScopeNode(scope Scope, node INode) {
	if scope == nil || scope != nil && scope.isTopScope() {
		return
	}
	scope.removeScopeNode(node)
}

func maybeAddScopeNode(scope Scope, node INode) {
	node.Node().createdIn = scope
	if scope != nil && scope.isTopScope() {
		return
	}
	scope.addScopeNode(node)
}

func updateNodeScope(scope Scope, node INode) {
	maybeRemoveScopeNode(node.Node().createdIn, node)
	maybeAddScopeNode(scope, node)
}

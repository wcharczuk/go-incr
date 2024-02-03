package incr

import "context"

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
func WithinBindScope[A INode](ctx context.Context, node A) A {
	if value := ctx.Value(bindScopeKey{}); value != nil {
		if scope, ok := value.(*bindScope); ok {
			addNodeToBindScope(ctx, node, scope)
		}
	}
	return node
}

type bindScopeKey struct{}

func withBindScope(ctx context.Context, scope *bindScope) context.Context {
	return context.WithValue(ctx, bindScopeKey{}, scope)
}

type bindScope struct {
	lhs      INode
	bind     INode
	rhsNodes *nodeList
}

func addNodeToBindScope(ctx context.Context, node INode, scope *bindScope) {
	TracePrintf(ctx, "%v adding to bound scope %v", scope.bind, node)
	node.Node().createdIn = scope
	scope.rhsNodes.Push(node)
}

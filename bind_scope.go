package incr

import "context"

type bindScopeKey struct{}

func withBindScope(ctx context.Context, scope *bindScope) context.Context {
	return context.WithValue(ctx, bindScopeKey{}, scope)
}

type bindScope struct {
	lhs      INode
	bind     INode
	rhsNodes *nodeList
}

// WithBindScope is a hack because go can't have nice things.
func WithBindScope[A INode](ctx context.Context, node A) A {
	if value := ctx.Value(bindScopeKey{}); value != nil {
		if scope, ok := value.(*bindScope); ok {
			addNodeToBindScope(ctx, node, scope)
		}
	}
	return node
}

func addNodeToBindScope(ctx context.Context, node INode, scope *bindScope) {
	node.Node().createdIn[scope.bind.Node().id] = scope
	scope.rhsNodes.Push(node)
}

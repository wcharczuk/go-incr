package incr

import (
	"context"
	"time"
)

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

// Root is is the top level bind scope.
func Root() *BindScope {
	return _root
}

func withBindScope(ctx context.Context, scope *BindScope) *BindScope {
	scope.inner = ctx
	return scope
}

func addNodeToBindScope(scope *BindScope, node INode) {
	if scope == nil || scope.root {
		return
	}
	TracePrintf(scope, "%v adding to bound scope %v", scope.bind, node)
	node.Node().createdIn = scope
	scope.rhsNodes.Push(node)
}

var _root = &BindScope{root: true}

// BindScope is the scope that nodes are created in.
//
// Its either nil or the most recent bind.
type BindScope struct {
	root     bool
	inner    context.Context
	lhs      INode
	bind     INode
	rhsNodes *nodeList
}

func (bs *BindScope) Deadline() (deadline time.Time, ok bool) {
	if bs.inner != nil {
		deadline, ok = bs.inner.Deadline()
	}
	return
}

func (bs *BindScope) Done() (done <-chan struct{}) {
	if bs.inner != nil {
		done = bs.inner.Done()
	}
	return
}

func (bs *BindScope) Err() (err error) {
	if bs.inner != nil {
		err = bs.inner.Err()
	}
	return
}

func (bs *BindScope) Value(key any) (value any) {
	if bs.inner != nil {
		value = bs.inner.Value(key)
	}
	return
}

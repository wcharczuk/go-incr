package incr

import "context"

// Adapters that let tests install a bare function as one of a node's cached
// optional-interface delegates.
//
// The node stores these as interfaces rather than bound method values so that
// creating a node does not allocate a closure per delegate; these wrappers keep
// the tests able to supply an inline function.

type stabilizeFunc func(context.Context) error

func (f stabilizeFunc) Node() *Node                         { return nil }
func (f stabilizeFunc) Stabilize(ctx context.Context) error { return f(ctx) }

type cutoffFunc func(context.Context) (bool, error)

func (f cutoffFunc) Node() *Node                              { return nil }
func (f cutoffFunc) Cutoff(ctx context.Context) (bool, error) { return f(ctx) }

type staleFunc func() bool

func (f staleFunc) Node() *Node { return nil }
func (f staleFunc) Stale() bool { return f() }

type shouldBeInvalidatedFunc func() bool

func (f shouldBeInvalidatedFunc) Node() *Node               { return nil }
func (f shouldBeInvalidatedFunc) ShouldBeInvalidated() bool { return f() }

type parentsFunc func() []INode

func (f parentsFunc) Node() *Node      { return nil }
func (f parentsFunc) Parents() []INode { return f() }

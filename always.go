package incr

// Always returns an incremental that is always stale and will be
// marked for recomputation.
func Always[A any](scope Scope, input Incr[A]) Incr[A] {
	return WithinScope(scope, &alwaysIncr[A]{
		n:       NewNode("always"),
		input:   input,
		parents: []INode{input},
	})
}

// AlwaysIncr is a type that implements the always stale incremental.
type AlwaysIncr[A any] interface {
	Incr[A]
	IAlways
}

type alwaysIncr[A any] struct {
	n       *Node
	input   Incr[A]
	parents []INode
}

func (a *alwaysIncr[A]) Parents() []INode {
	return a.parents
}

func (a *alwaysIncr[A]) Stale() bool {
	return true
}

func (a *alwaysIncr[A]) Always() {}

func (a *alwaysIncr[A]) Value() A {
	return a.input.Value()
}

func (a *alwaysIncr[A]) Node() *Node { return a.n }

func (a *alwaysIncr[A]) String() string {
	return a.n.String()
}

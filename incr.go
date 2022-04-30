package incr

type Incr[T comparable] interface {
	Node() *Node
	Value() T
}

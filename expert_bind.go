package incr

// ExpertBind returns a reference to an expert interface
// for bind nodes to enable working with the interior
// details of how Binds are implemented.
//
// Note there are no compatibility guarantees on this interface
// and you should use this interface at your own caution.
func ExpertBind[A any](b BindIncr[A]) IExpertBind {
	return expertBind(b)
}

func expertBind(b any) IExpertBind {
	typed, _ := b.(IExpertBind)
	return typed
}

type IExpertBind interface {
	// BindChange returns the
	BindChange() INode
	// Bound returns the right-hand-side of the node
	Bound() INode
}

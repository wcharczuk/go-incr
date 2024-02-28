package naive

func Map[A, B any](fn MapFn[A, B], inputs ...Node[A]) Node[B] {
	return &mapNodeImpl[A, B]{
		inputs: inputs,
		action: fn,
	}
}

type MapFn[A, B any] func(...A) B

type mapNodeImpl[A, B any] struct {
	inputs []Node[A]
	action MapFn[A, B]
}

func (n mapNodeImpl[A, B]) Value() B {
	inputs := make([]A, len(n.inputs))
	for x := 0; x < len(n.inputs); x++ {
		inputs[x] = n.inputs[x].Value()
	}
	return n.action(inputs...)
}

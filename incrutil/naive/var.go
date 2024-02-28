package naive

type VarNode[A any] interface {
	Node[A]
	SetValue(A)
}

func Var[A any](v A) VarNode[A] {
	return &varNodeImpl[A]{
		value: v,
	}
}

type varNodeImpl[A any] struct {
	value A
}

func (v *varNodeImpl[A]) Value() A { return v.value }

func (v *varNodeImpl[A]) SetValue(newValue A) {
	v.value = newValue
}

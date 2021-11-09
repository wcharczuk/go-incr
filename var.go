package incremental

// Var holds a value.
//
// It is typically mutable and used as an input
// for a computation.
type Var[A any] struct {
	value A
}

// Set sets the value.
func (v *Var[A]) Set(value A) {
	v.value = value
}

// Value implements Incr.
func (v Var[A]) Value() A {
	return v.value
}

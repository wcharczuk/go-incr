package incr

// IExpertVar are methods implemented by ExpertVar.
type IExpertVar[A any] interface {
	// SetValue allows you to set the underlying value of a var
	// _without_ marking it as stale or needing to be recomputed.
	//
	// This can be useful when deserializing graphs from some other state.
	SetValue(A)
}

// ExpertVar returns an "expert" version of a var node.
func ExpertVar[A any](v VarIncr[A]) IExpertVar[A] {
	return &expertVar[A]{v: v.(*varIncr[A])}
}

type expertVar[A any] struct {
	v *varIncr[A]
}

func (ev *expertVar[A]) SetValue(v A) {
	ev.v.value = v
}

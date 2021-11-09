package incremental

// Incr is the base interface.
type Incr[A any] interface {
	Value() A
}

// IncrStabilizer is a node that can be stabilized.
type IncrStabilizer interface {
	Stabilize()
}

// IncrFunc is a function that implements the base interface.
type IncrFunc[A any] func() A

// Value implements Incr.
func (i IncrFunc[A]) Value() A {
	return i()
}

// IncrFunc is a function that implements the base interface as
// a bind, or dynamic value provider.
type IncrBindFunc[A any] func() Incr[A]

// Value implements Incr.
func (ib IncrBindFunc[A]) Value() A {
	return ib().Value()
}

// Return creates a new incr from a given value.
//
// You can think of this as a constant.
func Return[A any](value A) Incr[A] {
	return IncrFunc[A](func() A { return value })
}

// Map returns the result of a given function `fn` on a given input.
func Map[A, B any](input Incr[A], fn func(A) B) Incr[B] {
	return IncrFunc[B](func() B {
		return fn(input.Value())
	})
}

// Map2 yields the application of two inputs, `inputA` and `inputB` yielding
// a 3rd type, C.
func Map2[A, B, C any](inputA Incr[A], inputB Incr[B], fn func(A, B) C) Incr[C] {
	return IncrFunc[C](func() C {
		return fn(inputA.Value(), inputB.Value())
	})
}

// Map3 can be made in terms of map2 but lets be honest I don't have time.
func Map3[A, B, C, D any] (inputA Incr[A], inputB Incr[B], inputC Incr[C], fn func(A, B, C) D) Incr[D] {
	return IncrFunc[D](func() D {
		return fn(inputA.Value(), inputB.Value(), inputC.Value())
	})
}

// MapIf returns one value or the other as the result of a given boolean incremental.
func MapIf[A any] (inputA, inputB Incr[A], c Incr[bool]) Incr[A] {
	return IncrFunc[A](func() A {
		if c.Value() {
			return inputA.Value()
		}
		return inputB.Value()
	})
}

// Map returns the result of a given function `fn` on a given input.
func Bind[A, B any](input Incr[A], fn func(A) Incr[B]) Incr[B] {
	return IncrBindFunc[B](func() Incr[B] {
		return fn(input.Value())
	})
}

// Map returns the result of a given function `fn` on a given input.
func BindIf[A any](inputA Incr[A], inputB Incr[A], c Incr[bool]) Incr[A] {
	return IncrBindFunc[A](func() Incr[A] {
		if c.Value() {
			return inputA
		}
		return inputB
	})
}

// Stabilize calls `.Stabilize()` on a node if it
// implements the `Stabilizer` interface.
func Stabilize[A any](input Incr[A]) {
	if typed, ok := input.(IncrStabilizer); ok {
		typed.Stabilize()
	}
}

// OnUpdate calls a function if a given input is evaluated.
func OnUpdate[A any](input Incr[A], fn func(A)) Incr[A] {
	return IncrFunc[A](func() A {
		value := input.Value()
		fn(value)
		return value
	})
}

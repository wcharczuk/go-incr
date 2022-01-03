package incr

// OkValue returns just the value from a (A,bool) return.
func OkValue[A any](v A, ok bool) A {
	return v
}

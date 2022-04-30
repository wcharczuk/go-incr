package incr

func Map[A, B comparable](a Incr[A], fn func(a A) B) Incr[B] {
	return nil
}

func Map2[A, B, C comparable](a Incr[A], b Incr[B], fn func(A, B) C) Incr[C] {
	return nil
}

func Map3[A, B, C, D comparable](a Incr[A], b Incr[B], c Incr[C], fn func(A, B, C) D) Incr[D] {
	return nil
}

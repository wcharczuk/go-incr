package incr

import (
	"context"
	"time"
)

// Delay returns a node that will only recompute after a given
// delay since it last recomputed expressed as a duration.
func Delay[A any](i Incr[A], delay time.Duration) Incr[A] {
	di := &delayIncr[A]{
		i:     i,
		delay: delay,
		now:   time.Now,
	}
	di.n = NewNode(
		di,
		OptNodeChildOf(i),
	)
	return di
}

type delayIncr[A any] struct {
	n     *Node
	delay time.Duration
	last  time.Time
	now   func() time.Time
	i     Incr[A]
}

func (di *delayIncr[A]) Value() A {
	return di.i.Value()
}

func (di *delayIncr[A]) Stale() bool { return false }

func (di *delayIncr[A]) Stabilize(ctx context.Context) error {
	now := di.now()
	if now.Sub(di.last) > di.delay {
		di.last = now
		return di.i.Stabilize(ctx)
	}
	return nil
}

func (di *delayIncr[A]) Node() *Node { return di.n }

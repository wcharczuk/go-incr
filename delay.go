package incr

import (
	"context"
	"time"
)

// Delay returns a node that will only recompute after a given
// delay since it last recomputed expressed as a duration.
func Delay[A comparable](i Incr[A], delay time.Duration) Incr[A] {
	di := &delayIncr[A]{
		i:     i,
		delay: delay,
	}
	di.n = NewNode(
		di,
		OptNodeChildOf(i),
	)
	return di
}

type delayIncr[A comparable] struct {
	n           *Node
	delay       time.Duration
	last        time.Time
	nowProvider func() time.Time
	i           Incr[A]
}

func (di *delayIncr[A]) now() time.Time {
	if di.nowProvider != nil {
		return di.nowProvider()
	}
	return time.Now()
}

func (di *delayIncr[A]) Value() A {
	return di.i.Value()
}

func (di *delayIncr[A]) Stabilize(ctx context.Context, g Generation) error {
	now := di.now()
	if now.Sub(di.last) > di.delay {
		di.last = now
		di.n.changedAt = g
		return di.i.Stabilize(ctx, g)
	}
	return nil
}

func (di *delayIncr[A]) Node() *Node { return di.n }

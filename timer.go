package incr

import (
	"context"
	"fmt"
	"time"
)

// Timer returns a special node type that fires if a given duration
// has elapsed since it last stabilized.
//
// When it stabilizes, it assumes the value of the input node, and causes
// any children (i.e. nodes that take the timer as input) to recompute if this
// is the first stabilization or if the timer has elapsed.
func Timer[A any](scope Scope, input Incr[A], every time.Duration) Incr[A] {
	return WithinScope(scope, &timerIncr[A]{
		n:           NewNode(KindTimer),
		clockSource: func(_ context.Context) time.Time { return time.Now().UTC() },
		every:       every,
		input:       input,
	})
}

var (
	_ Incr[struct{}] = (*timerIncr[struct{}])(nil)
	_ IAlways        = (*timerIncr[struct{}])(nil)
	_ ICutoff        = (*timerIncr[struct{}])(nil)
	_ IStabilize     = (*timerIncr[struct{}])(nil)
	_ fmt.Stringer   = (*timerIncr[struct{}])(nil)
)

type timerIncr[A any] struct {
	n           *Node
	clockSource func(context.Context) time.Time
	last        time.Time
	every       time.Duration
	input       Incr[A]
	value       A
}

func (ti *timerIncr[A]) Parents() []INode {
	return []INode{ti.input}
}

func (ti *timerIncr[A]) Node() *Node { return ti.n }

func (ti *timerIncr[A]) Value() A { return ti.value }

func (ti *timerIncr[A]) Always() {}

func (ti *timerIncr[A]) Cutoff(ctx context.Context) (bool, error) {
	now := ti.clockSource(ctx)
	return now.Sub(ti.last) < ti.every, nil
}

func (ti *timerIncr[A]) Stabilize(ctx context.Context) error {
	ti.last = ti.clockSource(ctx)
	ti.value = ti.input.Value()
	return nil
}

func (ti *timerIncr[A]) String() string {
	return ti.n.String()
}

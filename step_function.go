package incr

import (
	"context"
	"fmt"
	"slices"
	"time"
)

// Step is one change in a [StepFunction]: the value it takes from a given time.
type Step[A any] struct {
	At    time.Time
	Value A
}

// StepFunction returns an incremental that holds a value which changes at known times.
//
// It is the way to express something scheduled -- a rate that changes at market open, a
// limit that relaxes overnight, a feature that turns on at a announced moment -- as an
// input to a computation rather than as something a caller has to remember to poll. The
// value is whichever step's time has most recently passed, and initial applies before the
// first of them.
//
// Steps may be given in any order; they are sorted. Two steps at the same time are
// resolved in favor of the one given later, on the grounds that a caller listing both
// probably means the second to win.
//
// The node is woken once per step rather than continuously, so a step function with a
// handful of changes costs nothing between them however long they are apart.
func StepFunction[A any](scope Scope, clock *Clock, initial A, steps ...Step[A]) Incr[A] {
	ordered := make([]Step[A], len(steps))
	copy(ordered, steps)
	slices.SortStableFunc(ordered, func(a, b Step[A]) int { return a.At.Compare(b.At) })

	s := &stepFunctionIncr[A]{
		clock:   clock,
		initial: initial,
		steps:   ordered,
		value:   initial,
	}
	s.n = NewNode(KindStepFunction)
	WithinScope(scope, s)
	// wake for the first step that has not already passed; nothing to wake for if they
	// have all gone by, since the value will not change again
	s.entry = clock.register(s, s.nextBoundary(clock.Now()))
	return s
}

var (
	_ Incr[int]    = (*stepFunctionIncr[int])(nil)
	_ IStabilize   = (*stepFunctionIncr[int])(nil)
	_ fmt.Stringer = (*stepFunctionIncr[int])(nil)
)

type stepFunctionIncr[A any] struct {
	n       *Node
	clock   *Clock
	initial A
	steps   []Step[A]
	entry   *clockEntry
	value   A
}

func (s *stepFunctionIncr[A]) Node() *Node { return s.n }

func (s *stepFunctionIncr[A]) Value() A { return s.value }

// nextBoundary returns the time of the first step after now, or the zero time if the
// steps are exhausted.
func (s *stepFunctionIncr[A]) nextBoundary(now time.Time) time.Time {
	for _, step := range s.steps {
		if step.At.After(now) {
			return step.At
		}
	}
	return time.Time{}
}

func (s *stepFunctionIncr[A]) Stabilize(_ context.Context) error {
	now := s.clock.Now()
	// the value is the last step to have started, which is found by scanning rather than
	// by remembering a position: a clock that jumps forward skips steps, and the answer
	// has to be the same either way.
	value := s.initial
	for _, step := range s.steps {
		if step.At.After(now) {
			break
		}
		value = step.Value
	}
	s.value = value
	s.clock.rearm(s.entry, s.nextBoundary(now))
	return nil
}

func (s *stepFunctionIncr[A]) String() string { return s.n.String() }

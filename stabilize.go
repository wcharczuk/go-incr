package incr

import (
	"context"
)

// Stabilize stabilizes a computation.
func Stabilize(ctx context.Context, outputs ...Stabilizer) error {
	stale, err := stabilizeDiscoverStale(ctx, outputs)
	if err != nil {
		return err
	}
	if err = stabilize(ctx, stale); err != nil {
		return err
	}
	return nil
}

func stabilizeDiscoverStale(ctx context.Context, outputs []Stabilizer) (*Queue[Stabilizer], error) {
	seen := make(map[nodeID]bool)
	stale := new(Queue[Stabilizer])
	for _, o := range outputs {
		if err := stabilizeDiscoverStaleTraverse(ctx, seen, stale, o); err != nil {
			return nil, err
		}
	}
	return stale, nil
}

func stabilizeDiscoverStaleTraverse(ctx context.Context, seen map[nodeID]bool, stale *Queue[Stabilizer], s Stabilizer) error {
	sNode := s.getNode()
	sNodeID := sNode.id
	if _, ok := seen[sNodeID]; ok {
		return ErrStabilizeCycle
	}
	sStale := s.Stale()
	seen[sNodeID] = sStale
	if sStale {
		stale.Push(s)
	}
	for _, p := range s.getNode().parents {
		if err := stabilizeDiscoverStaleTraverse(ctx, seen, stale, p); err != nil {
			return err
		}
	}
	return nil
}

func stabilize(ctx context.Context, stale *Queue[Stabilizer]) error {
	return stale.ReverseEach(func(s Stabilizer) error {
		return s.Stabilize(ctx)
	})
}

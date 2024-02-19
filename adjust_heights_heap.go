package incr

import (
	"fmt"
	"sync"
)

func newAdjustHeightsHeap(maxHeightAllowed int) *adjustHeightsHeap {
	return &adjustHeightsHeap{
		nodesByHeight:    make([]*queue[INode], maxHeightAllowed),
		heightLowerBound: maxHeightAllowed + 1,
	}
}

type adjustHeightsHeap struct {
	mu               sync.Mutex
	nodesByHeight    []*queue[INode]
	numNodes         int
	maxHeightSeen    int
	heightLowerBound int
}

func (ah *adjustHeightsHeap) len() int {
	return ah.numNodes
}

func (ah *adjustHeightsHeap) maxHeightAllowed() int {
	return len(ah.nodesByHeight) - 1
}

func (ah *adjustHeightsHeap) setHeight(node INode, height int) error {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	return ah.setHeightUnsafe(node, height)
}

func (ah *adjustHeightsHeap) adjustHeights(rh *recomputeHeap, originalChild, originalParent INode) error {
	ah.mu.Lock()
	rh.mu.Lock()
	defer ah.mu.Unlock()
	defer rh.mu.Unlock()

	ah.heightLowerBound = originalChild.Node().height
	if err := ah.ensureHeightRequirementUnsafe(originalChild, originalParent, originalChild, originalParent); err != nil {
		return err
	}
	for ah.numNodes > 0 {
		parent, _ := ah.removeMinUnsafe()
		if parent.Node().heightInRecomputeHeap != HeightUnset {
			rh.fixUnsafe(parent)
		}
		for _, child := range parent.Node().children {
			if err := ah.ensureHeightRequirementUnsafe(originalChild, originalParent, child, parent); err != nil {
				return err
			}
		}
		if typed, typedOK := parent.(IBindChange); typedOK {
			for _, nodeOnRight := range typed.RightScopeNodes() {
				if nodeOnRight.Node().isNecessary() {
					if err := ah.ensureHeightRequirementUnsafe(originalChild, originalParent, nodeOnRight, parent); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (ah *adjustHeightsHeap) ensureHeightRequirementUnsafe(originalChild, originalParent, child, parent INode) error {
	if originalParent.Node().id == child.Node().id {
		return fmt.Errorf("cycle detected at %v to %v", originalChild, originalParent)
	}
	if parent.Node().height >= child.Node().height {
		// we set `child.height` after adding `child` to the heap, so that `child` goes
		// in the heap with its pre-adjusted height.
		ah.addUnsafe(child)
		if err := ah.setHeightUnsafe(child, parent.Node().height+1); err != nil {
			return err
		}
	}
	return nil
}

func (ah *adjustHeightsHeap) removeMinUnsafe() (node INode, ok bool) {
	if ah.numNodes == 0 {
		return
	}
	// NOTE (wc): we cannot start at heightLowerBound because of
	// transient parallel recomputation issues. We use zero here to avoid
	// locking the node metadata.
	for x := 0; x <= ah.maxHeightSeen; x++ {
		if ah.nodesByHeight[x] != nil && ah.nodesByHeight[x].len() > 0 {
			node, ok = ah.nodesByHeight[x].pop()
			ah.heightLowerBound = x
			node.Node().heightInAdjustHeightsHeap = HeightUnset
			ah.numNodes--
			return
		}
	}
	return
}

func (ah *adjustHeightsHeap) addUnsafe(node INode) {
	if node.Node().heightInAdjustHeightsHeap != HeightUnset {
		return
	}
	height := node.Node().height
	node.Node().heightInAdjustHeightsHeap = height
	if ah.nodesByHeight[height] == nil {
		ah.nodesByHeight[height] = new(queue[INode])
	}
	if height > ah.maxHeightSeen {
		ah.maxHeightSeen = height
	}
	ah.nodesByHeight[height].push(node)
	ah.numNodes++
}

func (ah *adjustHeightsHeap) setHeightUnsafe(node INode, height int) error {
	if height > ah.maxHeightAllowed() {
		return fmt.Errorf("cannot set node height above %d", ah.maxHeightAllowed())
	}
	if height > ah.maxHeightSeen {
		ah.maxHeightSeen = height
	}
	node.Node().height = height
	return nil
}

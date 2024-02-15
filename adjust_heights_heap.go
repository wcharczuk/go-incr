package incr

import (
	"fmt"
	"sync"
)

func newAdjustHeightsHeap(maxHeightAllowed int) *adjustHeightsHeap {
	return &adjustHeightsHeap{
		nodesByHeight: make([]*list[Identifier, INode], maxHeightAllowed),
		lookup:        make(set[Identifier]),
	}
}

type adjustHeightsHeap struct {
	mu               sync.Mutex
	nodesByHeight    []*list[Identifier, INode]
	lookup           set[Identifier]
	maxHeightSeen    int
	heightLowerBound int
}

func (ah *adjustHeightsHeap) len() (out int) {
	ah.mu.Lock()
	out = len(ah.lookup)
	ah.mu.Unlock()
	return
}

func (ah *adjustHeightsHeap) maxHeightAllowed() int {
	return len(ah.nodesByHeight) - 1
}

func (ah *adjustHeightsHeap) remove(node INode) {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	ah.removeUnsafe(node)
}

func (ah *adjustHeightsHeap) removeUnsafe(node INode) {
	if len(ah.lookup) == 0 {
		return
	}
	if !ah.lookup.has(node.Node().id) {
		return
	}
	delete(ah.lookup, node.Node().id)
	height := node.Node().heightInAdjustHeightsHeap
	if height >= ah.heightLowerBound && height <= ah.maxHeightSeen {
		if ah.nodesByHeight[height] != nil {
			ah.nodesByHeight[height].remove(node.Node().id)
		}
	}
}

const heightUnset = -1

func (ah *adjustHeightsHeap) add(node INode) {
	if node.Node().heightInAdjustHeightsHeap != heightUnset {
		return
	}
	height := node.Node().height
	node.Node().heightInAdjustHeightsHeap = height
	if ah.nodesByHeight[height] == nil {
		ah.nodesByHeight[height] = new(list[Identifier, INode])
	}
	if height > ah.maxHeightSeen {
		ah.maxHeightSeen = height
	}
	ah.nodesByHeight[height].push(node.Node().id, node)
	ah.lookup.add(node.Node().id)
}

func (ah *adjustHeightsHeap) removeMin() (node INode, ok bool) {
	if len(ah.lookup) == 0 {
		return
	}
	for x := 0; x <= ah.maxHeightSeen; x++ {
		if ah.nodesByHeight[x] != nil && ah.nodesByHeight[x].len() > 0 {
			var id Identifier
			id, node, ok = ah.nodesByHeight[x].pop()
			ah.heightLowerBound = x
			node.Node().heightInAdjustHeightsHeap = heightUnset
			delete(ah.lookup, id)
			return
		}
	}
	return
}

func (ah *adjustHeightsHeap) setHeight(node INode, height int) error {
	if height > ah.maxHeightAllowed() {
		return fmt.Errorf("cannot set node height above %d", ah.maxHeightAllowed())
	}
	if height > ah.maxHeightSeen {
		ah.maxHeightSeen = height
	}
	node.Node().height = height
	return nil
}

func (ah *adjustHeightsHeap) ensureHeightRequirement(originalChild, originalParent, child, parent INode) error {
	if originalParent.Node().id == child.Node().id {
		return fmt.Errorf("cycle detected at %v to %v", originalChild, originalParent)
	}
	if parent.Node().height >= child.Node().height {
		// we set `child.height` after adding `child` to the heap, so that `child` goes
		// in the heap with its pre-adjusted height.
		ah.add(child)
		if err := ah.setHeight(child, parent.Node().height+1); err != nil {
			return err
		}
	}
	return nil
}

func (ah *adjustHeightsHeap) adjustHeights(rh *recomputeHeap, originalChild, originalParent INode) error {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	if err := ah.ensureHeightRequirement(originalChild, originalParent, originalChild, originalParent); err != nil {
		return err
	}
	for len(ah.lookup) > 0 {
		parent, _ := ah.removeMin()
		// if the node is in the recompute heap
		if parent.Node().heightInRecomputeHeap >= 0 {
			rh.fix(parent)
		}
		for _, child := range parent.Node().children {
			if err := ah.ensureHeightRequirement(originalChild, originalParent, child, parent); err != nil {
				return err
			}
		}
		if typed, typedOK := parent.(IBindChange); typedOK {
			for _, nodeOnRight := range typed.RightScopeNodes() {
				if nodeOnRight.Node().isNecessary() {
					if err := ah.ensureHeightRequirement(originalChild, originalParent, nodeOnRight, parent); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

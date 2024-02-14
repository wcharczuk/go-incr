package incr

import (
	"fmt"
	"sync"
)

func newAdjustHeightsHeap(maxHeightAllowed int) *adjustHeightsHeap {
	return &adjustHeightsHeap{
		nodesByHeight: make([]map[Identifier]INode, maxHeightAllowed),
		lookup:        make(set[Identifier]),
	}
}

type adjustHeightsHeap struct {
	mu               sync.Mutex
	nodesByHeight    []map[Identifier]INode
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
			delete(ah.nodesByHeight[height], node.Node().id)
		}
	}
}

func (ah *adjustHeightsHeap) add(node INode) {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	ah.removeUnsafe(node)
	ah.addUnsafe(node)
}

func (ah *adjustHeightsHeap) addUnsafe(node INode) {
	ah.lookup.add(node.Node().id)
	height := node.Node().height
	node.Node().heightInAdjustHeightsHeap = height
	if ah.nodesByHeight[height] == nil {
		ah.nodesByHeight[height] = make(map[Identifier]INode)
	}
	if height > ah.maxHeightSeen {
		ah.maxHeightSeen = height
	}
	ah.nodesByHeight[height][node.Node().id] = node
}

func (ah *adjustHeightsHeap) removeMinUnsafe() (node INode, ok bool) {
	if len(ah.lookup) == 0 {
		return
	}
	for x := ah.heightLowerBound; x <= ah.maxHeightSeen; x++ {
		if ah.nodesByHeight[x] != nil && len(ah.nodesByHeight[x]) > 0 {
			node, _ = popMap(ah.nodesByHeight[x])
			ok = true
			ah.heightLowerBound = x
			delete(ah.lookup, node.Node().id)
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
	if height < ah.heightLowerBound {
		ah.heightLowerBound = height
	}
	node.Node().height = height
	return nil
}

func (ah *adjustHeightsHeap) ensureHeightRequirement(originalChild, originalParent, child, parent INode) error {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	return ah.ensureHeightRequirementUnsafe(originalChild, originalParent, child, parent)
}

func (ah *adjustHeightsHeap) ensureHeightRequirementUnsafe(originalChild, originalParent, child, parent INode) error {
	if originalParent.Node().id == child.Node().id {
		return fmt.Errorf("cycle detected at %v", child)
	}
	if parent.Node().height >= child.Node().height {
		if err := ah.setHeight(child, parent.Node().height+1); err != nil {
			return err
		}
		ah.removeUnsafe(child)
		ah.addUnsafe(child)
	}
	return nil
}

func (ah *adjustHeightsHeap) adjustHeights(rh *recomputeHeap, originalChild, originalParent INode) error {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	if err := ah.ensureHeightRequirementUnsafe(originalChild, originalParent, originalChild, originalParent); err != nil {
		return err
	}
	for len(ah.lookup) > 0 {
		node, _ := ah.removeMinUnsafe()
		rh.fix(node.Node().id)
		for _, child := range node.Node().children {
			if err := ah.ensureHeightRequirementUnsafe(originalChild, originalParent, child, node); err != nil {
				return err
			}
		}
		if typed, typedOK := node.(IBindChange); typedOK {
			for _, nodeOnRight := range typed.RightScopeNodes() {
				if nodeOnRight.Node().isNecessary() {
					if err := ah.ensureHeightRequirementUnsafe(originalChild, originalParent, node, nodeOnRight); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

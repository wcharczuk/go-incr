package incr

import (
	"fmt"
	"sync"
)

const (
	heightUnset = -1
)

func newAdjustHeightsHeap(maxHeightAllowed int) *adjustHeightsHeap {
	return &adjustHeightsHeap{
		nodesByHeight: make([]*queue[INode], maxHeightAllowed+32),
		lookup:        make(set[Identifier]),
	}
}

type adjustHeightsHeap struct {
	mu               sync.Mutex
	nodesByHeight    []*queue[INode]
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

func (ah *adjustHeightsHeap) isEmpty() bool {
	return ah.len() == 0
}

func (ah *adjustHeightsHeap) maxHeightAllowed() int {
	return len(ah.nodesByHeight)
}

func (ah *adjustHeightsHeap) setMaxHeightAllowed(height int) {
	if height < ah.maxHeightSeen {
		return
	}
	ah.nodesByHeight = make([]*queue[INode], height)
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
	if height > ah.heightLowerBound && height < ah.maxHeightSeen {
		if ah.nodesByHeight[height] != nil {
			ah.nodesByHeight[height].filter(func(qn INode) bool {
				return qn.Node().id != node.Node().id
			})
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
		ah.nodesByHeight[height] = new(queue[INode])
	}
	ah.nodesByHeight[height].push(node)
}

func (ah *adjustHeightsHeap) removeMin() (node INode, ok bool) {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	if len(ah.lookup) == 0 {
		return
	}
	for x := 0; x <= ah.maxHeightSeen; x++ {
		if ah.nodesByHeight[x] != nil && ah.nodesByHeight[x].len() > 0 {
			node, _ = ah.nodesByHeight[x].pop()
			node.Node().heightInAdjustHeightsHeap = heightUnset
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

func (ah *adjustHeightsHeap) ensureHeightRequirement(child, parent INode) error {
	if parent.Node().height >= child.Node().height {
		if err := ah.setHeight(child, parent.Node().height+1); err != nil {
			return err
		}
		ah.add(child)
	}
	return nil
}

func (ah *adjustHeightsHeap) adjustHeights(rh *recomputeHeap) error {
	for ah.len() > 0 {
		node, _ := ah.removeMin()
		rh.fix(node.Node().id)
		for _, child := range node.Node().children {
			if err := ah.ensureHeightRequirement(child, node); err != nil {
				return err
			}
		}
		if typed, ok := node.(IBind); ok {
			scope := typed.Scope()
			for _, nodeOnRight := range scope.rhsNodes {
				if node.Node().graph.isNecessary(nodeOnRight) {
					if err := ah.ensureHeightRequirement(node, nodeOnRight); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

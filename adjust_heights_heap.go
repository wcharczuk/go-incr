package incr

import "fmt"

const (
	heightUnset = -1
)

func newAdjustHeightsHeap(maxHeightAllowed int) *adjustHeightsHeap {
	return &adjustHeightsHeap{
		nodesByHeight: make([][]INode, maxHeightAllowed),
	}
}

type adjustHeightsHeap struct {
	nodesByHeight    [][]INode
	maxHeightSeen    int
	heightLowerBound int
	length           int
}

func (ah *adjustHeightsHeap) isEmpty() bool {
	return ah.length == 0
}

func (ah *adjustHeightsHeap) maxHeightAllowed() int {
	return len(ah.nodesByHeight)
}

func (ah *adjustHeightsHeap) setMaxHeightAllowed(height int) {
	if height < ah.maxHeightSeen {
		return
	}
	ah.nodesByHeight = make([][]INode, height)
}

func (ah *adjustHeightsHeap) remove(node INode) {
	if ah.isEmpty() {
		return
	}
	height := node.Node().heightInAdjustHeightsHeap
	if height > ah.heightLowerBound && height < ah.maxHeightSeen {
		ah.nodesByHeight[height] = remove(ah.nodesByHeight[height], node.Node().id)
	}
}

func (ah *adjustHeightsHeap) add(node INode) {
	if node.Node().heightInAdjustHeightsHeap == heightUnset {
		height := node.Node().height
		node.Node().heightInAdjustHeightsHeap = height
		ah.length++
		ah.nodesByHeight[height] = append(ah.nodesByHeight[height], node)
	}
}

func (ah *adjustHeightsHeap) removeMin() (node INode, ok bool) {
	if ah.isEmpty() {
		return
	}
	for x := ah.heightLowerBound; x < ah.maxHeightSeen; x++ {
		if len(ah.nodesByHeight[x]) > 0 {
			node, ah.nodesByHeight[x] = removeFirst(ah.nodesByHeight[x])
			node.Node().heightInAdjustHeightsHeap = heightUnset
			ok = true
			ah.heightLowerBound = x
			ah.length--
		}
	}
	return
}

func (ah *adjustHeightsHeap) setHeight(node INode, height int) error {
	if height > ah.maxHeightSeen {
		ah.maxHeightSeen = height
		if height > ah.maxHeightAllowed() {
			return fmt.Errorf("cannot set node height above %d", ah.maxHeightAllowed())
		}
	}
	node.Node().height = height
	return nil
}

func (ah *adjustHeightsHeap) ensureHeightRequirement(originalChild, originalParent, child, parent INode) error {
	if parent.Node().id == originalChild.Node().id {
		return fmt.Errorf("cycle detected at %v", parent)
	}
	if child.Node().height >= parent.Node().height {
		ah.add(parent)
		return ah.setHeight(parent, child.Node().height+1)
	}
	return nil
}

func (ah *adjustHeightsHeap) adjustHeights(rh *recomputeHeap, originalChild, originalParent INode) error {
	if err := ah.ensureHeightRequirement(
		originalChild, originalParent,
		originalChild, originalParent,
	); err != nil {
		return err
	}
	for ah.length > 0 {
		child, _ := ah.removeMin()
		rh.increaseHeightUnsafe(child)
		for _, parent := range child.Node().parents {
			if err := ah.ensureHeightRequirement(originalChild, originalParent, child, parent); err != nil {
				return err
			}
		}
		if typed, ok := child.(IBind); ok {
			scope := typed.Scope()
			for _, nodeOnRight := range scope.Nodes() {
				if nodeOnRight.Node().isNecessary() {
					if err := ah.ensureHeightRequirement(originalChild, originalParent, child, nodeOnRight); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

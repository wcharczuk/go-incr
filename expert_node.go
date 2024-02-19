package incr

import (
	"reflect"
)

// ExpertNode returns an "expert" interface to interact with nodes.
//
// Note there are no compatibility guarantees on this interface
// and you should use this interface at your own caution.
func ExpertNode(in INode) IExpertNode {
	return &expertNode{incr: in, node: in.Node()}
}

// IExpertNode is an expert interface for nodes.
//
// Note there are no compatibility guarantees on this interface
// and you should use this interface at your own caution.
type IExpertNode interface {
	CreatedIn() Scope
	SetCreatedIn(Scope)
	SetID(Identifier)

	Valid() bool
	SetValid(bool)

	Height() int
	SetHeight(int)

	HeightInRecomputeHeap() int
	SetHeightInRecomputeHeap(int)

	HeightInAdjustHeightsHeap() int
	SetHeightInAdjustHeightsHeap(int)

	ChangedAt() uint64
	SetChangedAt(uint64)
	SetAt() uint64
	SetSetAt(uint64)
	RecomputedAt() uint64
	SetRecomputedAt(uint64)

	IsNecessary() bool
	IsStale() bool
	IsInRecomputeHeap() bool

	Always() bool
	SetAlways(bool)

	Observer() bool
	SetObserver(bool)

	Children() []INode
	AddChildren(...INode)
	Parents() []INode
	AddParents(...INode)
	Observers() []IObserver
	AddObservers(...IObserver)
	RemoveChild(Identifier)
	RemoveParent(Identifier)
	RemoveObserver(Identifier)

	// ComputePseudoHeight walks the node graph up from a given node
	// computing the height of the node in-respect to its full graph.
	//
	// This is in contrast to how the node's height is calculated during
	// computation, which is only when it is linked or bound, and only
	// in respect to its immediate parents.
	//
	// This method is useful in advanced scenarios where you may be
	// rebuilding a graph from scratch dynamically.
	ComputePseudoHeight() int

	// Value returns the underlying value of the node
	// as an untyped `interface{}` for use in debugging.
	Value() any
}

type expertNode struct {
	incr INode
	node *Node
}

func (en *expertNode) CreatedIn() Scope         { return en.node.createdIn }
func (en *expertNode) SetCreatedIn(scope Scope) { en.node.createdIn = scope }

func (en *expertNode) SetID(id Identifier) {
	en.node.id = id
}

func (en *expertNode) Valid() bool {
	return en.node.valid
}

func (en *expertNode) SetValid(valid bool) {
	en.node.valid = valid
}

func (en *expertNode) Height() int { return en.node.height }

func (en *expertNode) SetHeight(height int) {
	en.node.height = height
}

func (en *expertNode) HeightInRecomputeHeap() int { return en.node.heightInRecomputeHeap }

func (en *expertNode) SetHeightInRecomputeHeap(height int) {
	en.node.heightInRecomputeHeap = height
}

func (en *expertNode) HeightInAdjustHeightsHeap() int { return en.node.heightInAdjustHeightsHeap }

func (en *expertNode) SetHeightInAdjustHeightsHeap(height int) {
	en.node.heightInAdjustHeightsHeap = height
}

func (en *expertNode) ChangedAt() uint64 { return en.node.changedAt }

func (en *expertNode) SetChangedAt(changedAt uint64) {
	en.node.changedAt = changedAt
}

func (en *expertNode) SetAt() uint64 { return en.node.setAt }

func (en *expertNode) SetSetAt(setAt uint64) {
	en.node.setAt = setAt
}

func (en *expertNode) RecomputedAt() uint64 { return en.node.recomputedAt }

func (en *expertNode) SetRecomputedAt(recomputedAt uint64) {
	en.node.recomputedAt = recomputedAt
}

func (en *expertNode) IsNecessary() bool       { return en.node.isNecessary() }
func (en *expertNode) IsStale() bool           { return en.node.isStale() }
func (en *expertNode) IsInRecomputeHeap() bool { return en.node.heightInRecomputeHeap != HeightUnset }

func (en *expertNode) Always() bool { return en.node.always }
func (en *expertNode) SetAlways(always bool) {
	en.node.always = always
}

func (en *expertNode) Observer() bool { return en.node.observer }
func (en *expertNode) SetObserver(observer bool) {
	en.node.observer = observer
}

func (en *expertNode) AddChildren(c ...INode) {
	en.node.addChildren(c...)
}

func (en *expertNode) Children() []INode {
	return en.node.children
}

func (en *expertNode) AddParents(c ...INode) {
	en.node.addParents(c...)
}

func (en *expertNode) Parents() []INode {
	return en.node.parents
}

func (en *expertNode) AddObservers(o ...IObserver) {
	en.node.addObservers(o...)
}

func (en *expertNode) Observers() []IObserver {
	return en.node.observers
}

func (en *expertNode) RemoveChild(id Identifier) {
	en.node.removeChild(id)
}

func (en *expertNode) RemoveParent(id Identifier) {
	en.node.removeParent(id)
}

func (en *expertNode) RemoveObserver(id Identifier) {
	en.node.removeObserver(id)
}

func (en *expertNode) Value() any {
	rv := reflect.ValueOf(en.incr)
	valueMethod := rv.MethodByName("Value")
	if !valueMethod.IsValid() {
		return nil
	}
	res := valueMethod.Call(nil)
	if len(res) > 0 {
		return res[0].Interface()
	}
	return nil
}

func (en *expertNode) ComputePseudoHeight() int {
	return en.computePseudoHeightCached(make(map[Identifier]int), en.incr)
}

func (en *expertNode) computePseudoHeightCached(cache map[Identifier]int, n INode) int {
	nn := n.Node()
	if height, ok := cache[nn.ID()]; ok {
		return height
	}

	var maxParentHeight int
	for _, p := range nn.parents {
		parentHeight := en.computePseudoHeightCached(cache, p)
		if parentHeight > maxParentHeight {
			maxParentHeight = parentHeight
		}
	}
	var finalHeight int
	if nn.height > maxParentHeight {
		finalHeight = nn.height
	} else {
		finalHeight = maxParentHeight + 1
	}
	cache[nn.ID()] = finalHeight
	return finalHeight
}

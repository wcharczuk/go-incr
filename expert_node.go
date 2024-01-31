package incr

import "reflect"

// ExpertNode returns an "expert" interface to interact with nodes.
//
// USE AT YOUR OWN RISK.
func ExpertNode(in INode) IExpertNode {
	return &expertNode{incr: in, node: in.Node()}
}

// IExpertNode is an expert interface for nodes.
type IExpertNode interface {
	SetID(Identifier)
	Height() int
	SetHeight(int)
	ChangedAt() uint64
	SetChangedAt(uint64)
	SetAt() uint64
	SetSetAt(uint64)
	RecomputedAt() uint64
	SetRecomputedAt(uint64)
	BoundAt() uint64
	SetBoundAt(uint64)
	Always() bool
	SetAlways(bool)

	AddChildren(...INode)
	AddParents(...INode)
	RemoveChild(Identifier)
	RemoveParent(Identifier)

	// Value returns the underlying value of the node
	// as an untyped `interface{}` for use in debugging.
	Value() any
}

type expertNode struct {
	incr INode
	node *Node
}

func (en *expertNode) SetID(id Identifier) {
	en.node.id = id
}

func (en *expertNode) Height() int { return en.node.height }

func (en *expertNode) SetHeight(height int) {
	en.node.height = height
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

func (en *expertNode) BoundAt() uint64 { return en.node.boundAt }

func (en *expertNode) SetBoundAt(boundAt uint64) {
	en.node.boundAt = boundAt
}

func (en *expertNode) Always() bool { return en.node.always }

func (en *expertNode) SetAlways(always bool) {
	en.node.always = always
}

func (en *expertNode) AddChildren(c ...INode) {
	en.node.addChildren(c...)
}

func (en *expertNode) AddParents(c ...INode) {
	en.node.addParents(c...)
}

func (en *expertNode) RemoveChild(id Identifier) {
	en.node.removeChild(id)
}

func (en *expertNode) RemoveParent(id Identifier) {
	en.node.removeParent(id)
}

func (en *expertNode) Value() any {
	rv := reflect.ValueOf(en.incr)
	valueMethod := rv.MethodByName("Value")
	res := valueMethod.Call(nil)
	if len(res) > 0 {
		return res[0].Interface()
	}
	return nil
}

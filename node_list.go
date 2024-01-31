package incr

import "sync"

func newNodeList(nodes ...INode) *nodeList {
	nl := &nodeList{
		list: make(map[Identifier]INode),
	}
	nl.PushUnsafe(nodes...)
	return nl
}

type nodeList struct {
	mu   sync.Mutex
	list map[Identifier]INode
}

func (nl *nodeList) Lock() {
	nl.mu.Lock()
}

func (nl *nodeList) Unlock() {
	nl.mu.Unlock()
}

func (nl *nodeList) Len() (length int) {
	nl.mu.Lock()
	length = len(nl.list)
	nl.mu.Unlock()
	return
}

func (nl *nodeList) IsEmpty() bool {
	return nl.Len() == 0
}

func (nl *nodeList) IsEmptyUnsafe() bool {
	return len(nl.list) == 0
}

func (nl *nodeList) Push(nodes ...INode) {
	nl.Lock()
	defer nl.Unlock()

	nl.PushUnsafe(nodes...)
}

func (nl *nodeList) PushUnsafe(nodes ...INode) {
	for _, n := range nodes {
		nl.list[n.Node().id] = n
	}
}

func (nl *nodeList) Remove(n INode) {
	nl.Lock()
	defer nl.Unlock()
	delete(nl.list, n.Node().id)
}

func (nl *nodeList) RemoveKey(id Identifier) {
	nl.Lock()
	defer nl.Unlock()

	delete(nl.list, id)
}

func (nl *nodeList) HasKey(id Identifier) (ok bool) {
	nl.Lock()
	defer nl.Unlock()

	_, ok = nl.list[id]
	return
}

func (nl *nodeList) Has(n INode) (ok bool) {
	nl.Lock()
	defer nl.Unlock()

	_, ok = nl.list[n.Node().id]
	return
}

func (nl *nodeList) HasUnsafe(n INode) (ok bool) {
	_, ok = nl.list[n.Node().id]
	return
}

func (nl *nodeList) PopAllUnsafe() (out []INode) {
	out = make([]INode, 0, len(nl.list))
	for _, n := range nl.list {
		out = append(out, n)
	}
	clear(nl.list)
	return
}

func (nl *nodeList) ConsumeEach(fn func(INode)) {
	nl.Lock()
	defer nl.Unlock()

	for _, n := range nl.list {
		fn(n)
	}
	clear(nl.list)
}

func (nl *nodeList) Each(fn func(INode)) {
	nl.Lock()
	defer nl.Unlock()
	for _, n := range nl.list {
		fn(n)
	}
}

func (nl *nodeList) Values() (out []INode) {
	nl.Lock()
	defer nl.Unlock()
	out = make([]INode, 0, len(nl.list))
	for _, n := range nl.list {
		out = append(out, n)
	}
	return
}

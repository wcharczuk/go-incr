package incr

func newNodeList(nodes ...INode) *nodeList {
	nl := &nodeList{
		list: new(list[Identifier, INode]),
	}
	for _, n := range nodes {
		nl.PushUnsafe(n)
	}
	return nl
}

type nodeList struct {
	list *list[Identifier, INode]
}

func (nl *nodeList) Lock() {
	nl.list.mu.Lock()
}

func (nl *nodeList) Unlock() {
	nl.list.mu.Unlock()
}

func (nl *nodeList) Clear() {
	nl.list.Clear()
}

func (nl *nodeList) Len() int {
	return nl.list.Len()
}

func (nl *nodeList) IsEmpty() bool {
	return nl.list.IsEmpty()
}

func (nl *nodeList) IsEmptyUnsafe() bool {
	return nl.list.isEmptyUnsafe()
}

func (nl *nodeList) Push(nodes ...INode) {
	nl.Lock()
	defer nl.Unlock()
	nl.PushUnsafe(nodes...)
}

func (nl *nodeList) PushUnsafe(nodes ...INode) {
	for _, n := range nodes {
		nl.list.pushUnsafe(n.Node().id, n)
	}
}

func (nl *nodeList) Remove(n INode) {
	nl.list.Remove(n.Node().id)
}

func (nl *nodeList) RemoveUnsafe(n INode) {
	nl.list.removeUnsafe(n.Node().id)
}

func (nl *nodeList) RemoveKey(id Identifier) {
	nl.list.Remove(id)
}

func (nl *nodeList) RemoveKeyUnsafe(id Identifier) {
	nl.list.removeUnsafe(id)
}

func (nl *nodeList) HasKey(id Identifier) (ok bool) {
	ok = nl.list.Has(id)
	return
}

func (nl *nodeList) HasKeyUnsafe(id Identifier) (ok bool) {
	ok = nl.list.hasUnsafe(id)
	return
}

func (nl *nodeList) Has(n INode) (ok bool) {
	ok = nl.list.Has(n.Node().id)
	return
}

func (nl *nodeList) HasUnsafe(n INode) (ok bool) {
	ok = nl.list.hasUnsafe(n.Node().id)
	return
}

func (nl *nodeList) Pop() (out INode, ok bool) {
	_, out, ok = nl.list.Pop()
	return
}

func (nl *nodeList) PopUnsafe() (out INode, ok bool) {
	_, out, ok = nl.list.popUnsafe()
	return
}

func (nl *nodeList) PopAll() (out []INode) {
	out = nl.list.PopAll()
	return
}

func (nl *nodeList) PopAllUnsafe() (out []INode) {
	out = nl.list.popAllUnsafe()
	return
}

func (nl *nodeList) ConsumeEach(fn func(INode)) {
	nl.list.ConsumeEach(fn)
}

func (nl *nodeList) Each(fn func(INode) error) error {
	return nl.list.Each(fn)
}

func (nl *nodeList) Values() (out []INode) {
	out = nl.list.Values()
	return
}

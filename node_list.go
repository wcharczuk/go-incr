package incr

func newNodeList(nodes ...INode) *nodeList {
	nl := &nodeList{
		list: new(list[Identifier, INode]),
	}
	nl.PushUnsafe(nodes...)
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
	nl.Lock()
	defer nl.Unlock()

	nl.list.removeUnsafe(n.Node().id)
}

func (nl *nodeList) RemoveKey(id Identifier) {
	nl.Lock()
	defer nl.Unlock()

	nl.list.removeUnsafe(id)
}

func (nl *nodeList) HasKey(id Identifier) (ok bool) {
	nl.Lock()
	defer nl.Unlock()

	ok = nl.list.hasUnsafe(id)
	return
}

func (nl *nodeList) Has(n INode) (ok bool) {
	nl.Lock()
	defer nl.Unlock()

	ok = nl.list.hasUnsafe(n.Node().id)
	return
}

func (nl *nodeList) HasUnsafe(n INode) (ok bool) {
	ok = nl.list.hasUnsafe(n.Node().id)
	return
}

func (nl *nodeList) PopAllUnsafe() (out []INode) {
	out = nl.list.popAllUnsafe()
	return
}

func (nl *nodeList) ConsumeEach(fn func(INode)) {
	nl.Lock()
	defer nl.Unlock()

	nl.list.consumeEachUnsafe(fn)
}

func (nl *nodeList) Each(fn func(INode)) {
	nl.Lock()
	defer nl.Unlock()

	nl.list.eachUnsafe(fn)
}

func (nl *nodeList) Values() (out []INode) {
	nl.Lock()
	defer nl.Unlock()
	out = nl.list.valuesUnsafe()
	return
}

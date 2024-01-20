package incr

import "slices"

func newNodeLookup(nodes ...INode) nodeLookup {
	output := make(nodeLookup, len(nodes))
	output.Add(nodes...)
	return output
}

// nodeLookup is a lookup for nodes ...
type nodeLookup map[Identifier]INode

// Add adds a list of values to the lookup.
func (nl nodeLookup) Add(values ...INode) {
	for _, v := range values {
		nl[v.Node().ID()] = v
	}
}

func (nl nodeLookup) Values() []INode {
	values := make([]INode, 0, len(nl))
	for _, v := range nl {
		values = append(values, v)
	}
	slices.SortStableFunc(values, nodeSorter)
	return values
}

// Remove removes a value.
func (nl nodeLookup) Remove(v INode) {
	delete(nl, v.Node().ID())
}

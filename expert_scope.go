package incr

// ExpertScope returns an "expert" interface to access or
// modify internal fields of the Scope type.
//
// Note there are no compatibility guarantees on this interface
// and you should use this interface at your own risk.
func ExpertScope(scope Scope) IExpertScope {
	return &expertScope{scope: scope}
}

type IExpertScope interface {
	// IsTopScope returns if a given scope is the top scope, e.g. it represents the graph.
	IsTopScope() bool
	// IsScopeValid returns if a given scope is valid, that is if it's still included in the graph.
	IsScopeValid() bool
	// IsScopeNecessary returns if a given scope is observed, that is if it's part of the graph output.
	IsScopeNecessary() bool
	// ScopeGraph returns the top level scope, i.e. the root graph.
	ScopeGraph() *Graph
	// ScopeHeight returns the effective height of the scope as derrived by it's root node.
	ScopeHeight() int
	// AddScopeNode adds a node to a scope.
	AddScopeNode(INode)
	// NewIdentifier returns a new identifier from the graph configured identifier provider.
	NewIdentifier() Identifier
}

type expertScope struct {
	scope Scope
}

func (es expertScope) IsTopScope() bool {
	return es.scope.isTopScope()
}

func (es expertScope) IsScopeValid() bool {
	return es.scope.isScopeValid()
}

func (es expertScope) IsScopeNecessary() bool {
	return es.scope.isScopeNecessary()
}

func (es expertScope) ScopeGraph() *Graph {
	return es.scope.scopeGraph()
}

func (es expertScope) ScopeHeight() int {
	return es.scope.scopeHeight()
}

func (es expertScope) AddScopeNode(n INode) {
	es.scope.addScopeNode(n)
}

func (es expertScope) NewIdentifier() Identifier {
	return es.scope.newIdentifier()
}

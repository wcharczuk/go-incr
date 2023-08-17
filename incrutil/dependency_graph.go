package incrutil

import (
	"context"

	"github.com/wcharczuk/go-incr"
)

// DependencyGraph is a list of dependencies that
// can be resolved incrementally using an action.
type DependencyGraph[Result any] struct {
	Dependencies []Dependency
	Action       incr.MapNFunc[Result, Result]
	OnUpdate     func(context.Context)
}

// Dependency is pseudo-dependency metadata.
type Dependency struct {
	Name      string
	DependsOn []string
}

// Create walks the dependency graph and returns the "leaves" of the graph, or the nodes that
// are not depended on by any other nodes.
func (dg DependencyGraph[Result]) Create() *incr.Graph {
	dependencyLookup := dg.createDependencyLookup()

	// build "dependedBy" list(s)
	for _, p := range dg.Dependencies {
		for _, d := range p.DependsOn {
			dependencyLookup[d].dependedBy = append(dependencyLookup[d].dependedBy, p.Name)
		}
	}

	// build package incrementals
	// including the relationships between the
	// packages and their dependencies.
	packageIncrementals := dg.createDependencyIncrLookup()

	// build "leaves" list by filtering for
	// nodes with zero "dependedBy" entries
	var leaves []incr.INode
	for _, d := range dependencyLookup {
		if len(d.dependedBy) == 0 {
			leaves = append(leaves, packageIncrementals[d.Name])
		}
	}

	return incr.New(leaves...)
}

func (dg DependencyGraph[Result]) createDependencyLookup() (output map[string]*dependencyWithDependedBy) {
	output = make(map[string]*dependencyWithDependedBy)
	for index := range dg.Dependencies {
		output[dg.Dependencies[index].Name] = &dependencyWithDependedBy{Dependency: dg.Dependencies[index]}
	}
	return
}

func (dg DependencyGraph[Result]) createDependencyIncrLookup() (output map[string]DependencyIncr[Result]) {
	output = make(map[string]DependencyIncr[Result])
	for _, d := range dg.Dependencies {
		output[d.Name] = dg.createDependencyIncr(d)
	}
	for _, p := range dg.Dependencies {
		for _, d := range p.DependsOn {
			output[p.Name].AddInput(output[d])
		}
	}
	return
}

func (dg DependencyGraph[Result]) createDependencyIncr(d Dependency) DependencyIncr[Result] {
	output := incr.MapN[Result, Result](dg.Action)
	if dg.OnUpdate != nil {
		output.Node().OnUpdate(dg.OnUpdate)
	}
	output.Node().SetLabel(d.Name)
	return output
}

type dependencyWithDependedBy struct {
	Dependency
	dependedBy []string
}

// DependencyIncr is a dependency graph node that takes
// potentially many results and returns a result itself by calling
// the dependency action on those input results.
type DependencyIncr[Result any] incr.MapNIncr[Result, Result]

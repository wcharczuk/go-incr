package incrutil

import (
	"context"

	"github.com/wcharczuk/go-incr"
)

// DependencyGraph is a list of dependencies that
// can be resolved incrementally using a action function.
//
// The goal with DependencyGraph is to show how you can build up
// abstractions above incremental, e.g. how you'd represent a specific
// type of directed graph.
type DependencyGraph[Result any] struct {
	Dependencies []Dependency
	Action       func(context.Context, Dependency) (Result, error)
	OnUpdate     func(context.Context, Dependency)
}

// Dependency is pseudo-dependency metadata.
type Dependency struct {
	Name      string
	DependsOn []string
}

// Create walks the dependency graph and returns the "leaves" of the graph, or the nodes that
// are not depended on by any other nodes.
func (dg DependencyGraph[Result]) Create(ctx context.Context) (*incr.Graph, map[string]DependencyIncr[Result], error) {
	dependencyLookup := dg.createDependencyLookup()

	// build "dependedBy" list(s)
	for _, p := range dg.Dependencies {
		for _, d := range p.DependsOn {
			dependencyLookup[d].dependedBy = append(dependencyLookup[d].dependedBy, p.Name)
		}
	}

	graph := incr.New()
	// build package incrementals
	// including the relationships between the
	// packages and their dependencies.
	packageIncrementals, err := dg.createDependencyIncrLookup(ctx, graph)
	if err != nil {
		return nil, nil, err
	}

	// build "leaves" list by filtering for
	// nodes with zero "dependedBy" entries
	var leaves []DependencyIncr[Result]
	for _, d := range dependencyLookup {
		if len(d.dependedBy) == 0 {
			leaves = append(leaves, packageIncrementals[d.Name])
		}
	}
	for _, n := range leaves {
		_ = incr.Observe[Result](graph, n)
	}
	return graph, packageIncrementals, nil
}

func (dg DependencyGraph[Result]) createDependencyLookup() (output map[string]*dependencyWithDependedBy) {
	output = make(map[string]*dependencyWithDependedBy)
	for index := range dg.Dependencies {
		output[dg.Dependencies[index].Name] = &dependencyWithDependedBy{Dependency: dg.Dependencies[index]}
	}
	return
}

func (dg DependencyGraph[Result]) createDependencyIncrLookup(ctx context.Context, graph *incr.Graph) (output map[string]DependencyIncr[Result], err error) {
	output = make(map[string]DependencyIncr[Result])
	for _, d := range dg.Dependencies {
		output[d.Name] = dg.createDependencyIncr(graph, d)
	}
	for _, p := range dg.Dependencies {
		for _, d := range p.DependsOn {
			if err = output[p.Name].AddInput(output[d]); err != nil {
				return
			}
		}
	}
	return
}

func (dg DependencyGraph[Result]) mapAction(d Dependency) func(ctx context.Context, _ ...Result) (Result, error) {
	return func(ctx context.Context, _ ...Result) (Result, error) {
		return dg.Action(ctx, d)
	}
}

func (dg DependencyGraph[Result]) mapOnUpdate(d Dependency) func(context.Context) {
	return func(ctx context.Context) {
		dg.OnUpdate(ctx, d)
	}
}

func (dg DependencyGraph[Result]) createDependencyIncr(graph *incr.Graph, d Dependency) DependencyIncr[Result] {
	output := incr.MapNContext[Result, Result](graph, dg.mapAction(d))
	if dg.OnUpdate != nil {
		output.Node().OnUpdate(dg.mapOnUpdate(d))
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

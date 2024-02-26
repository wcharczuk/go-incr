package incrutil

import (
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
)

// DependencyGraph is a list of dependencies that
// can be resolved incrementally using a action function.
//
// The goal with DependencyGraph is to show how you can build up
// abstractions above incremental, e.g. how you'd represent a specific
// type of directed graph.
type DependencyGraph[Result any] struct {
	// Dependency is a flat list of individual dependency items, and their "parents" or
	// the other items they depend on.
	//
	// The dependency graph is then built from thist list using the `Create` function.
	Dependencies []Dependency

	// CheckIfStale is an optional delegate to provide an automatic method for determining
	// if a dependency needs to be rebuilt.
	//
	// If this is not provided, you will have to set the dependencies stale individually
	// with `incr.Graph::SetStale(Dependency)`, which is returned by the function `Create`.
	CheckIfStale func(context.Context, Dependency) (bool, error)

	// Action is the function that is called to "resolve" or build a dependency.
	//
	// It could also be thought of as the map or stabilize function.
	Action func(context.Context, Dependency) (Result, error)
}

// Dependency is pseudo-dependency metadata.
type Dependency struct {
	// Name is a unique name for the dependency, and should be referred to by the `DependsOn` slice.
	Name string

	// DependsOn is a slice of names of other dependencies that this dependency depends on.
	//
	// Specifically the dependencies named in this list will need to be built before this dependency.
	DependsOn []string
}

// Create walks the dependency graph and returns the "leaves" of the graph, or the nodes that
// are not depended on by any other nodes.
func (dg DependencyGraph[Result]) Create(ctx context.Context) (*incr.Graph, map[string]DependencyIncr[Result], error) {
	dependencyLookup := dg.createDependencyLookup()
	for _, p := range dg.Dependencies {
		for _, d := range p.DependsOn {
			if _, exists := dependencyLookup[d]; !exists {
				return nil, nil, fmt.Errorf("dependency graph; dependency %q names non-existent dependency %q", p.Name, d)
			}
			dependencyLookup[d].dependedBy = append(dependencyLookup[d].dependedBy, p.Name)
		}
	}
	graph := incr.New()
	packageIncrementals, err := dg.createDependencyIncrLookup(graph)
	if err != nil {
		return nil, nil, err
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

func (dg DependencyGraph[Result]) createDependencyIncrLookup(g *incr.Graph) (output map[string]DependencyIncr[Result], err error) {
	output = make(map[string]DependencyIncr[Result])
	for _, d := range dg.Dependencies {
		if _, alreadyExists := output[d.Name]; alreadyExists {
			err = fmt.Errorf("dependency graph; duplicate dependency %q", d.Name)
			return
		}
		output[d.Name], err = dg.createDependencyIncr(g, d)
		if err != nil {
			return
		}
	}

	for _, p := range dg.Dependencies {
		for _, d := range p.DependsOn {
			if _, exists := output[d]; !exists {
				err = fmt.Errorf("dependency graph; dependency %q names non-existent dependency %q", p.Name, d)
				return
			}
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

func (dg DependencyGraph[Result]) createDependencyIncr(g *incr.Graph, d Dependency) (DependencyIncr[Result], error) {
	output := incr.MapNContext[Result, Result](g, dg.mapAction(d))
	output.Node().SetLabel(d.Name)
	if dg.CheckIfStale != nil {
		_ = incr.SentinelContext(g, func(ctx context.Context) (bool, error) {
			return dg.CheckIfStale(ctx, d)
		}, output)
	}
	_, err := incr.Observe(g, output)
	return output, err
}

type dependencyWithDependedBy struct {
	Dependency
	dependedBy []string
}

// DependencyIncr is a dependency graph node that takes
// potentially many results and returns a result itself by calling
// the dependency action on those input results.
type DependencyIncr[Result any] incr.MapNIncr[Result, Result]

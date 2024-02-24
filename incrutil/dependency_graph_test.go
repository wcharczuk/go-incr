package incrutil

import (
	"context"
	"sync"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_DependencyGraph(t *testing.T) {
	ctx := testContext()

	var actionedMu sync.Mutex
	actioned := make(map[string]int)

	stale := []string{}

	dg := DependencyGraph[string]{
		Dependencies: []Dependency{
			{Name: "cmd/blazectl", DependsOn: []string{"pkg/config", "pkg/engine", "pkg/util"}},
			{Name: "cmd/blazesrv", DependsOn: []string{"pkg/config", "pkg/engine", "pkg/util"}},
			{Name: "pkg/config", DependsOn: []string{"pkg/util"}},
			{Name: "pkg/engine", DependsOn: []string{"pkg/config", "pkg/util"}},
			{Name: "pkg/util"},
		},
		Action: func(ctx context.Context, d Dependency) (string, error) {
			actionedMu.Lock()
			actioned[d.Name]++
			actionedMu.Unlock()
			return "ok!", nil
		},
		CheckIfStale: func(_ context.Context, d Dependency) (bool, error) {
			for _, s := range stale {
				if s == d.Name {
					return true, nil
				}
			}
			return false, nil
		},
	}

	graph, nodes, err := dg.Create(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 5, len(nodes))

	testutil.Equal(t, 0, actioned["cmd/blazectl"])
	testutil.Equal(t, 0, actioned["cmd/blazesrv"])
	testutil.Equal(t, 0, actioned["pkg/config"])
	testutil.Equal(t, 0, actioned["pkg/engine"])
	testutil.Equal(t, 0, actioned["pkg/util"])

	err = graph.ParallelStabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, 1, actioned["cmd/blazectl"])
	testutil.Equal(t, 1, actioned["cmd/blazesrv"])
	testutil.Equal(t, 1, actioned["pkg/config"])
	testutil.Equal(t, 1, actioned["pkg/engine"])
	testutil.Equal(t, 1, actioned["pkg/util"])

	stale = []string{"pkg/engine"}

	err = graph.ParallelStabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, 2, actioned["cmd/blazectl"])
	testutil.Equal(t, 2, actioned["cmd/blazesrv"])
	testutil.Equal(t, 1, actioned["pkg/config"])
	testutil.Equal(t, 2, actioned["pkg/engine"])
	testutil.Equal(t, 1, actioned["pkg/util"])
}

func Test_DependencyGraph_duplicateDependencyName(t *testing.T) {
	ctx := testContext()

	dg := DependencyGraph[string]{
		Dependencies: []Dependency{
			{Name: "cmd/blazectl", DependsOn: []string{"pkg/config", "pkg/engine", "pkg/util"}},
			{Name: "cmd/blazectl", DependsOn: []string{"pkg/config", "pkg/engine", "pkg/util"}},
			{Name: "pkg/config", DependsOn: []string{"pkg/util"}},
			{Name: "pkg/engine", DependsOn: []string{"pkg/config", "pkg/util"}},
			{Name: "pkg/util"},
		},
		Action: func(ctx context.Context, d Dependency) (string, error) {
			return "ok!", nil
		},
	}

	graph, nodes, err := dg.Create(ctx)
	testutil.Error(t, err)
	testutil.Equal(t, `dependency graph; duplicate dependency "cmd/blazectl"`, err.Error())
	testutil.Nil(t, graph)
	testutil.Equal(t, 0, len(nodes))
}

func Test_DependencyGraph_missingDependency(t *testing.T) {
	ctx := testContext()

	dg := DependencyGraph[string]{
		Dependencies: []Dependency{
			{Name: "cmd/blazectl", DependsOn: []string{"pkg/config", "pkg/engine", "pkg/util"}},
			{Name: "cmd/blazesrv", DependsOn: []string{"pkg/config", "not-a-thing", "pkg/util"}},
			{Name: "pkg/config", DependsOn: []string{"pkg/util"}},
			{Name: "pkg/engine", DependsOn: []string{"pkg/config", "pkg/util"}},
			{Name: "pkg/util"},
		},
		Action: func(ctx context.Context, d Dependency) (string, error) {
			return "ok!", nil
		},
	}

	graph, nodes, err := dg.Create(ctx)
	testutil.Error(t, err)
	testutil.Equal(t, `dependency graph; dependency "cmd/blazesrv" names non-existent dependency "not-a-thing"`, err.Error())
	testutil.Nil(t, graph)
	testutil.Equal(t, 0, len(nodes))
}

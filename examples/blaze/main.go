package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil"
)

type BuildResult struct {
	Package string
	Elapsed time.Duration
	Output  string
}

func main() {
	ctx := context.Background()
	if os.Getenv("DEBUG") != "" {
		ctx = incr.WithTracing(ctx)
	}

	dg := incrutil.DependencyGraph[BuildResult]{
		Dependencies: []incrutil.Dependency{
			{Name: "cmd/blazectl", DependsOn: []string{"pkg/config", "pkg/engine", "pkg/util"}},
			{Name: "cmd/blazesrv", DependsOn: []string{"pkg/config", "pkg/engine", "pkg/util"}},
			{Name: "pkg/config", DependsOn: []string{"pkg/util"}},
			{Name: "pkg/engine", DependsOn: []string{"pkg/config", "pkg/util"}},
			{Name: "pkg/util"},
		},
		OnUpdate: func(_ context.Context, d incrutil.Dependency) {
			fmt.Printf("built: %s\n", d.Name)
		},
		Action: func(ctx context.Context, d incrutil.Dependency) (BuildResult, error) {
			start := time.Now()
			var delay = time.Duration(250 + rand.Intn(1500))
			<-time.After(delay * time.Millisecond)
			elapsed := time.Since(start)
			return BuildResult{
				Package: d.Name,
				Elapsed: elapsed,
				Output:  fmt.Sprintf("%s built in %v", d.Name, elapsed),
			}, nil
		},
	}
	graph, nodes := dg.Create()

	// one caveat here; we're stabilizing all the leaves
	// but because they're connected through children, we end
	// up doing basically no-op stabilizations after the first
	// node is stabilized
	if err := graph.ParallelStabilize(ctx); err != nil {
		log.Printf("error: %v", err)
	}

	// in real world usage we would have some way to get fsnotify hints on files matching a
	// glob, which we would then use to trigger this SetStale call.
	graph.SetStale(nodes["pkg/engine"])
	fmt.Println("pkg/engine now invalid")

	if err := graph.ParallelStabilize(ctx); err != nil {
		log.Printf("error: %v", err)
	}
}

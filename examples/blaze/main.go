package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/wcharczuk/go-incr"
)

// Package is pseudo-package metadata.
type Package struct {
	name       string
	dependsOn  []string
	dependedBy []string
}

func createBuildGraph(packages ...Package) (leaves []incr.INode, packageIncrementals map[string]PackageBuildIncr) {
	packageLookup := createPackageLookup(packages...)

	// build "dependedBy" list(s)
	for _, p := range packages {
		for _, d := range p.dependsOn {
			packageLookup[d].dependedBy = append(packageLookup[d].dependedBy, p.name)
		}
	}

	// build package incrementals
	// including the relationships between the
	// packages and their dependencies.
	packageIncrementals = createPackageIncrementalLookup(packages...)

	// build "leaves" list by filtering to
	// nodes with zero "dependedBy" entries
	for _, p := range packageLookup {
		if len(p.dependedBy) == 0 {
			leaves = append(leaves, packageIncrementals[p.name])
		}
	}
	return
}

func createPackageLookup(packages ...Package) (output map[string]*Package) {
	output = make(map[string]*Package)
	for index := range packages {
		output[packages[index].name] = &packages[index]
	}
	return
}

type BuildResult struct {
	Package string
	Elapsed time.Duration
	Output  string
}

type PackageBuildIncr = incr.MapNIncr[BuildResult, BuildResult]

func createPackageIncrementalLookup(packages ...Package) (output map[string]PackageBuildIncr) {
	output = make(map[string]PackageBuildIncr)
	for _, p := range packages {
		output[p.name] = createBuildPackageIncr(p)
	}
	for _, p := range packages {
		for _, d := range p.dependsOn {
			output[p.name].AddInput(output[d])
		}
	}
	return
}

func createBuildPackageIncr(p Package) PackageBuildIncr {
	output := incr.MapN(buildPackageFunc(p))
	output.Node().OnUpdate(func(context.Context) {
		fmt.Printf("built: %s\n", p.name)
	})
	output.Node().SetLabel(p.name)
	return output
}

func buildPackageFunc(p Package) incr.MapNFunc[BuildResult, BuildResult] {
	return func(ctx context.Context, inputs ...BuildResult) (BuildResult, error) {
		start := time.Now()
		var delay = time.Duration(250 + rand.Intn(1500))
		<-time.After(delay * time.Millisecond)
		elapsed := time.Since(start)
		return BuildResult{
			Package: p.name,
			Elapsed: elapsed,
			Output:  fmt.Sprintf("%s built in %v", p.name, elapsed),
		}, nil
	}
}

func main() {
	ctx := incr.WithTracing(context.Background())
	packages := []Package{
		{name: "cmd/blazectl", dependsOn: []string{"pkg/config", "pkg/engine", "pkg/util"}},
		{name: "cmd/blazesrv", dependsOn: []string{"pkg/config", "pkg/engine", "pkg/util"}},
		{name: "pkg/config", dependsOn: []string{"pkg/util"}},
		{name: "pkg/engine", dependsOn: []string{"pkg/config", "pkg/util"}},
		{name: "pkg/util"},
	}
	nodes, lookup := createBuildGraph(packages...)

	// one caveat here; we're stabilizing all the leaves
	// but because they're connected through children, we end
	// up doing basically no-op stabilizations after the first
	// node is stabilized
	if err := incr.Stabilize(ctx, nodes...); err != nil {
		log.Printf("error: %v", err)
	}

	// in real world usage we would have some way to get fsnotify hints on files matching a
	// glob, which we would then use to trigger this SetStale call.
	incr.SetStale(lookup["pkg/engine"])

	if err := incr.Stabilize(ctx, nodes...); err != nil {
		log.Printf("error: %v", err)
	}
}

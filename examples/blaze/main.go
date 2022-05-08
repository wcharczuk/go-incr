package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wcharczuk/go-incr"
)

type Package struct {
	name       string
	dependsOn  []string
	dependedBy []string
}

func createBuildGraph(packages ...Package) []incr.INode {
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
	packageIncrementalLookup := createPackageIncrementalLookup(packages...)

	// build "output" list by filtering to
	// nodes with zero "dependedBy" entries
	var output []incr.INode
	for _, p := range packageLookup {
		if len(p.dependedBy) == 0 {
			output = append(output, packageIncrementalLookup[p.name])
		}
	}
	return output
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
		output[p.name] = incr.MapN(buildPackageFunc(p))
	}
	for _, p := range packages {
		for _, d := range p.dependsOn {
			output[p.name].AddInput(output[d])
		}
	}
	return
}

func buildPackageFunc(p Package) incr.MapNFunc[BuildResult, BuildResult] {
	return func(ctx context.Context, inputs ...BuildResult) (BuildResult, error) {
		start := time.Now()
		<-time.After(500 * time.Millisecond)
		elapsed := time.Since(start)
		return BuildResult{
			Package: p.name,
			Elapsed: elapsed,
			Output:  fmt.Sprintf("%s built in %v", p.name, elapsed),
		}, nil
	}
}

func createPackageIncr(p Package) incr.INode {
	return incr.MapN(buildPackageFunc(p))
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
	nodes := createBuildGraph(packages...)
	if err := incr.Stabilize(ctx, nodes...); err != nil {
		log.Printf("error: %v", err)
	}
}

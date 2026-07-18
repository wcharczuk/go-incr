// Command build_graph is an incremental build system: sources are compiled, their
// discovered imports become dependency edges, and a link step combines the results.
//
// A build system is close to the archetypal use for incremental computation, and it
// exercises two things a simple dependency graph does not:
//
//   - A cutoff on canonical content. Editing a file without changing what it means -- a
//     comment, whitespace -- must not rebuild anything downstream. The cutoff is what
//     turns "the file was written" into "the file changed", and it only works if every
//     downstream step reads the canonical form rather than the file itself.
//   - Dependencies discovered by reading the file. What a source depends on is not known
//     until it has been parsed, so the graph's shape comes from its own inputs. That is
//     what [incr.Bind] is for.
//
// The output reports which steps ran for each edit, which is the only thing that
// actually matters about a build system.
package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/wcharczuk/go-incr"
)

// source is a file's contents as the editor last wrote them.
type source struct {
	Name string
	Body string
}

// unit is a compiled source together with what it depended on.
type unit struct {
	Name         string
	Symbols      []string
	Dependencies int
}

func main() {
	ctx := context.Background()
	g := incr.New(incr.OptGraphMaxHeight(64))

	files := map[string]incr.VarIncr[source]{
		"util.go": incr.Var(g, source{Name: "util.go", Body: "package util\nfunc Helper() {}\n"}),
		"api.go": incr.Var(g, source{Name: "api.go", Body: "import util\n" +
			"package api\nfunc Handler() {}\n"}),
		"main.go": incr.Var(g, source{Name: "main.go", Body: "import api\n" +
			"package main\nfunc main() {}\n"}),
	}

	// counters, reset before each build so the report shows the work of that build only
	var canonicalized, parsed, compiled, linked int

	// A build step per file, wired so that api.go depends on util.go and main.go on
	// api.go -- but with those edges established from the imports found in the source
	// rather than declared here.
	units := make(map[string]incr.Incr[unit], len(files))
	var build func(name string) incr.Incr[unit]
	build = func(name string) incr.Incr[unit] {
		if existing, ok := units[name]; ok {
			return existing
		}
		file := files[name]

		// Reduce the file to a canonical form, then cut off on that. Everything
		// downstream reads the canonical form rather than the raw bytes, which is the
		// part that matters: if any downstream step took the file itself as an input it
		// would recompute whenever the file was written, and the cutoff would buy
		// nothing. A real build system would hash a parsed representation rather than
		// carry normalized text, but the shape is the same.
		canonical := incr.Map(g, file, func(s source) string {
			canonicalized++
			return normalize(s.Body)
		})
		stable := incr.CutoffEqual(g, canonical)

		// Parsing gives the file's imports, which are not known until it is read. Note
		// this reads stable, not file.
		imports := incr.Map(g, stable, func(body string) []string {
			parsed++
			return importsOf(body)
		})

		// The dependency edges come from that list, so the graph's shape depends on its
		// own inputs. Bind rebuilds this node's inputs whenever the import list changes,
		// and releases the edges it no longer needs.
		result := incr.Bind(g, imports, func(bs incr.Scope, deps []string) incr.Incr[unit] {
			dependencies := make([]incr.Incr[unit], 0, len(deps))
			for _, dep := range deps {
				if _, ok := files[dep+".go"]; ok {
					dependencies = append(dependencies, build(dep+".go"))
				}
			}
			// compile once the dependencies are available
			return incr.Map2(bs, stable, incr.All(bs, dependencies...),
				func(_ string, built []unit) unit {
					compiled++
					symbols := map[string]struct{}{name: {}}
					for _, b := range built {
						for _, symbol := range b.Symbols {
							symbols[symbol] = struct{}{}
						}
					}
					return unit{Name: name, Symbols: sortedKeys(symbols), Dependencies: len(deps)}
				})
		})
		units[name] = result
		return result
	}

	root := build("main.go")
	binary := incr.Map(g, root, func(u unit) string {
		linked++
		return fmt.Sprintf("binary(%s) main.go imports %d",
			strings.Join(u.Symbols, ","), u.Dependencies)
	})
	observed := incr.MustObserve(g, binary)

	report := func(what string) {
		canonicalized, parsed, compiled, linked = 0, 0, 0, 0
		if err := g.Stabilize(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%-34s canonicalized=%d parsed=%d compiled=%d linked=%d\n",
			what, canonicalized, parsed, compiled, linked)
	}

	report("initial build")
	fmt.Println("  ->", observed.Value())

	// A comment-only edit. The canonical form is unchanged, so the cutoff fires and
	// nothing past it runs at all.
	fmt.Println()
	files["util.go"].Update(func(s source) source {
		s.Body = "package util\n// a note for the reader\nfunc Helper() {}\n"
		return s
	})
	report("comment-only edit to util.go")

	// A real change to a leaf. Its own compile reruns, and so does everything that
	// depends on it -- but nothing that does not.
	fmt.Println()
	files["util.go"].Update(func(s source) source {
		s.Body = "package util\nfunc Helper() {}\nfunc Extra() {}\n"
		return s
	})
	report("real edit to util.go")
	fmt.Println("  ->", observed.Value())

	// Editing a leaf that nothing depends on touches only that file.
	fmt.Println()
	files["main.go"].Update(func(s source) source {
		s.Body = "import api\npackage main\nfunc main() { /* changed */ }\n"
		return s
	})
	report("edit to main.go only")

	// Adding an import changes the shape of the graph, not just its values: main.go now
	// depends on util.go directly as well as through api.go.
	fmt.Println()
	files["main.go"].Update(func(s source) source {
		s.Body = "import api\nimport util\npackage main\nfunc main() {}\n"
		return s
	})
	report("add an import to main.go")
	fmt.Println("  ->", observed.Value())

	// Removing it again releases the edge that is no longer needed.
	fmt.Println()
	files["main.go"].Update(func(s source) source {
		s.Body = "import api\npackage main\nfunc main() {}\n"
		return s
	})
	report("remove the import again")
	fmt.Println("  ->", observed.Value())
}

// normalize strips comments and blank space, so that an edit which does not change what
// the file means produces the same canonical form. The point is that the cutoff compares
// meaning rather than bytes.
func normalize(body string) string {
	var out []string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// sortedKeys returns a set's members in order, so the reported symbol list is stable.
func sortedKeys(in map[string]struct{}) []string {
	out := make([]string, 0, len(in))
	for key := range in {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

// importsOf returns the names imported by a source file.
func importsOf(body string) []string {
	var out []string
	for _, line := range strings.Split(body, "\n") {
		if name, ok := strings.CutPrefix(strings.TrimSpace(line), "import "); ok {
			out = append(out, strings.TrimSpace(name))
		}
	}
	sort.Strings(out)
	return out
}

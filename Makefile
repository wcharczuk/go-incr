.PHONY: all
all: test lint

.PHONY: test
test:
	@go test ./...

# What CI runs. The scaling measurements are opt-in and run separately, since timing
# them alongside the rest of the suite measures the suite as much as the code.
.PHONY: ci
ci: test-repeated test-race scaling fuzz-smoke

.PHONY: test-repeated
test-repeated:
	@go test -count 10 ./...

.PHONY: test-race
test-race:
	@go test -race -count 1 ./...

# A minute per target, which checks the targets still build and that nothing is broken
# outright. Finding new bugs needs far longer -- see `fuzz` below and CONTRIBUTING.md.
.PHONY: fuzz-smoke
fuzz-smoke:
	@go test -run '^$$' -fuzz 'FuzzGraph$$' -fuzztime 60s .
	@go test -run '^$$' -fuzz 'FuzzMap$$' -fuzztime 45s ./incrutil/pmap
	@go test -run '^$$' -fuzz 'FuzzSymmetricDiff$$' -fuzztime 45s ./incrutil/pmap

# A real campaign, for when changing anything structural: edges, heights, teardown, bind
# relinking, or the persistent tree's balancing. Override with e.g. FUZZTIME=2h.
FUZZTIME ?= 30m
.PHONY: fuzz
fuzz:
	@go test -run '^$$' -fuzz 'FuzzGraph$$' -fuzztime $(FUZZTIME) .

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: scaling
scaling:
	@INCR_SCALING_TESTS=1 go test -p 1 -count 1 -run 'Test_scaling|_scaling|_work' -v ./...

.PHONY: bench
bench:
	@go test -run=XXX -bench=.

.PHONY: bench-profile
bench-profile:
	@go test -run=XXX -bench=. -benchmem -cpuprofile profile.out
	@go tool pprof profile.out

# Compares this library against Jane Street's OCaml incremental; see _bench/README.md
# for the toolchain it needs.
.PHONY: bench-compare
bench-compare:
	@./_bench/run.sh

.PHONY: cover
cover:
	-go test -coverprofile=cover.out ./...
	-go tool cover -func=cover.out
	-@rm cover.out

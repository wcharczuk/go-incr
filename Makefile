.PHONY: bench
bench:
	@go test -run=XXX -bench=.

.PHONY: bench-profile-cpu 
bench-profile-cpu:
	@go test -run=XXX -bench=Benchmark_Stabilize_withPreInitialize_16384 -cpuprofile bench-cpu.out

.PHONY: cover
cover:
	-go test -v -coverprofile=cover.out
	-go tool cover -func=cover.out
	-@rm cover.out

bench:
	@go test -run=XXX -bench=.

bench-profile-cpu:
	@go test -run=XXX -bench=Benchmark_Stabilize_withPreInitialize_16384 -cpuprofile bench-cpu.out

cover:
	@go test -v -coverprofile=cover.out
	@go tool cover -func=cover.out

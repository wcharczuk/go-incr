
ensure-reflex:
	@go install github.com/cespare/reflex@latest

ensure-pprof:
	@go install github.com/google/pprof@latest

bench-profile:
	@go test -run=XXX -bench=. -cpuprofile bench-cpu.out

watch:
	@reflex -g *.go -- go test -timeout 1s
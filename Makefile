
ensure-reflex:
	@go install github.com/cespare/reflex@latest

ensure-pprof:
	@go install github.com/google/pprof@latest

bench:
	@go test -run=XXX -bench=.

bench-profile-cpu:
	@go test -run=XXX -bench=. -cpuprofile bench-cpu.out

watch-test:
	@reflex -g *.go -- go test -timeout 1s

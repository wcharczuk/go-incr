.PHONY: bench
bench:
	@go test -run=XXX -bench=.

.PHONY: bench-profile
bench-profile:
	@go test -run=XXX -bench=. -benchmem -cpuprofile profile.out
	@go tool pprof profile.out

.PHONY: cover
cover:
	-go test -v -coverprofile=cover.out
	-go tool cover -func=cover.out
	-@rm cover.out

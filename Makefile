
ensure-reflex:
	@go install github.com/cespare/reflex@latest

watch:
	@reflex -g *.go -- go test -timeout 1s
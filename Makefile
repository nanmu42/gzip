.PHONY: test bench benchrace tidy

test:
	@find . -name go.mod -execdir go test -race -coverprofile=coverage.txt -covermode=atomic \;

bench:
	@find . -name go.mod -execdir go test -benchmem -bench . \;

benchrace:
	@find . -name go.mod -execdir go test -race -benchmem -bench . \;

tidy:
	@find . -name go.mod -execdir go mod tidy \;
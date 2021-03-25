.PHONY: test bench benchrace tidy dep

test:
	@find . -name go.mod -execdir pwd \; | xargs -L1 -I '{}' bash -c 'cd {} && go test -race -coverprofile=coverage.txt -covermode=atomic ./...'

bench:
	@find . -name go.mod -execdir pwd \; | xargs -L1 -I '{}' bash -c 'cd {} && go test -benchmem -bench .'

benchrace:
	@find . -name go.mod -execdir pwd \; | xargs -L1 -I '{}' bash -c 'cd {} && go test -race -benchmem -bench .'

tidy:
	@find . -name go.mod -execdir pwd \; | xargs -L1 -I '{}' bash -c 'cd {} && go mod tidy'

dep:
	@find . -name go.mod -execdir pwd \; | xargs -L1 -I '{}' bash -c 'cd {} && go get ./...'
VERSION="`git describe --abbrev=0 --tags`"
COMMIT="`git rev-list -1 --abbrev-commit HEAD`"
TEST_PACKAGES=$(shell go list ./... | grep -v /examples/ | grep -v /expt | grep -v /repl)

all: clean fmt build test

fmt:
	@echo "Formatting..."
	@goimports -l -w ./

clean:
	@echo "Cleaning up..."
	@go mod tidy -v

test:
	@echo "Running tests..."
	@go test -cover $(TEST_PACKAGES)

test-verbose:
	@echo "Running tests..."
	@go test -v -cover ./...

benchmark:
	@echo "Running benchmarks..."
	@go test -benchmem -run="none" -bench="Benchmark.*" -v ./...

build:
	@echo "Running build..."
	@go build ./...

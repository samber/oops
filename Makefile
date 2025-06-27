MODULES=$(shell go list -m)
MODULE_DIRS=$(shell go list -m -f '{{.Dir}}')

build:
	go build -v ./...

test:
	go test -race -v ${MODULES} ./...
watch-test:
	reflex -t 50ms -s -- sh -c 'gotest -race -v ${MODULES} ./...'

bench:
	go test -benchmem -count 3 -bench ${MODULES} ./...
watch-bench:
	reflex -t 50ms -s -- sh -c 'go test -benchmem -count 3 -bench ${MODULES} ./...'

coverage:
	go test -v -coverprofile=cover.out -covermode=atomic ${MODULES} ./...
	go tool cover -html=cover.out -o cover.html

tools:
	go install github.com/cespare/reflex@latest
	go install github.com/rakyll/gotest@latest
	go install github.com/psampaz/go-mod-outdated@latest
	go install github.com/jondot/goweight@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go get -t -u golang.org/x/tools/cmd/cover
	go install github.com/sonatype-nexus-community/nancy@latest
	go install golang.org/x/perf/cmd/benchstat@latest
	go install github.com/cespare/prettybench@latest
	go mod tidy

lint:
	golangci-lint run --timeout 60s --max-same-issues 50 ${MODULE_DIRS}
lint-fix:
	golangci-lint run --timeout 60s --max-same-issues 50 --fix ${MODULE_DIRS}

audit:
	go list -json -m all | nancy sleuth

outdated:
	go list -u -m -json all | go-mod-outdated -update -direct

weight:
	goweight

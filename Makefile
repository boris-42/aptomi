GIT_VERSION=$(shell git describe --tags --long --dirty)
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-X github.com/Aptomi/aptomi/pkg/slinga/version.gitVersion=${GIT_VERSION} -X github.com/Aptomi/aptomi/pkg/slinga/version.gitCommit=${GIT_COMMIT} -X github.com/Aptomi/aptomi/pkg/slinga/version.buildDate=${BUILD_DATE}"

.PHONY: default
default: clean build test

.PHONY: vendor
vendor:
	glide install --strip-vendor

.PHONY: vendor-no-color
vendor-no-color:
	glide --no-color install --strip-vendor

.PHONY: profile
profile:
	@echo "Profiling CPU for 15 seconds"
	go test -bench . -benchtime 15s ./pkg/slinga/engine -cpuprofile cpu.out
	go tool pprof -web engine.test cpu.out

.PHONY: coverage
coverage:
	@echo "Calculating code coverage"
	echo 'mode: atomic' > coverage.out && go list ./pkg/... | xargs -n1 -I{} sh -c 'go test -short -covermode=atomic -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.out' && rm coverage.tmp
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test
test:
	go test -short -v ./cmd/...
	go test -short -v ./pkg/...
	@echo "\nAll unit tests passed"

.PHONY: alltest
alltest:
	go test -v ./cmd/...
	go test -v ./pkg/...
	@echo "\nAll unit & integration tests passed"

.PHONY: test-loop
test-loop:
	while go test -v ./pkg/...; do :; done

.PHONY: clean-run-noop
clean-run-noop:
	$(eval TMP := $(shell mktemp -d))
	APTOMI_DB=$(TMP) tools/demo-local-policy-init.sh

.PHONY: smoke
smoke: alltest install clean-run-noop
	-rm -f aptomi

.PHONY: build
build:
	CGO_ENABLED=0 go build ${LDFLAGS} -v -i -o aptomi github.com/Aptomi/aptomi/cmd/aptomi

.PHONY: install
install:
	CGO_ENABLED=0 go install -v ${LDFLAGS} github.com/Aptomi/aptomi/cmd/aptomi

.PHONY: fmt
fmt:
	go fmt $$(go list ./... | grep -v /vendor/)

.PHONY: vet
vet:
	go tool vet -all -shadow ./cmd ./pkg || echo "\nSome vet checks failed\n"

.PHONY: lint
lint:
	gometalinter --deadline=120s ./pkg/... ./cmd/... | grep -v 'should not use dot imports'

.PHONY: validate
validate: fmt vet lint
	@echo "\nAll validations passed"

.PHONY: clean
clean:
	-rm -f aptomi
	go clean -r -i

VERSION=$(shell git describe --always --dirty --abbrev=10)
LDFLAGS="-X github.com/canonical/lxd-cluster-manager/version.version=$(VERSION)"

.PHONY: default
default: build

# Build targets.
.PHONY: compile
compile: build compile-binary

.PHONY: compile-binary
compile-binary:
	go build -v ./cmd/lxd-cluster-mgr
	go build -v ./cmd/lxd-cluster-mgrd

.PHONY: build
build:
	cd ui && yarn install && yarn build
	rm -rf internal/api/static &>/dev/null
	cd internal/api && go generate
	go install -v \
		-ldflags $(LDFLAGS) \
		./cmd/lxd-cluster-mgr
	go install -v \
		-ldflags $(LDFLAGS) \
		./cmd/lxd-cluster-mgrd

# Testing targets.
.PHONY: test
test: test-static test-e2e

.PHONY: test-e2e
test-e2e:
	go test -count=1 -v ./test/e2e

.PHONY: test-static
test-static:
ifeq ($(shell command -v golangci-lint 2> /dev/null),)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
endif
	golangci-lint run --timeout 5m

# Cleanup temp and build artifacts.
.PHONY: clean
clean:
	rm -rf internal/api/static
	rm -rf state
	cd ui && yarn clean

# Update targets.
.PHONY: update-gomod
update-gomod:
	go get -u ./...
	go mod tidy

# Update lxd-generate generated database helpers.
.PHONY: update-schema
update-schema:
	go generate ./internal/database/...
	gofmt -s -w ./internal/database
	goimports -w ./internal/database
	@echo "Code generation completed"


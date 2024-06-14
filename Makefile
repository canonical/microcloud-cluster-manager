.PHONY: default
default: build

# Build targets.
.PHONY: compile
compile: build compile-binary

.PHONY: compile-binary
compile-binary:
	go build -v ./cmd/lxd-site-mgr
	go build -v ./cmd/lxd-site-mgrd

.PHONY: build
build:
	cd ui && yarn install && yarn build
	rm -rf internal/api/static &>/dev/null
	cd internal/api && go generate
	go install -v ./cmd/lxd-site-mgr
	go install -v ./cmd/lxd-site-mgrd

# Testing targets.
.PHONY: check
check: check-static check-unit check-system

.PHONY: check-unit
check-unit:
	go test ./...

.PHONY: check-system
check-system: build
	./test/main.sh

.PHONY: check-static
check-static:
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
	go generate ./...
	gofmt -s -w ./database/
	goimports -w ./database/
	@echo "Code generation completed"


.PHONY: default
default: build

# Build targets.
.PHONY: build
build:
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
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
endif
ifeq ($(shell command -v shellcheck 2> /dev/null),)
	echo "Please install shellcheck"
	exit 1
endif
ifeq ($(shell command -v revive 2> /dev/null),)
	go install github.com/mgechev/revive@latest
endif
	golangci-lint run --timeout 5m
	revive -set_exit_status ./...

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

# Dev targets.

# Start the daemon in development mode.
.PHONY: start-daemon-dev
start-daemon-dev:
	go run ./cmd/lxd-site-mgrd --state-dir state_dir_1 &
	go run ./cmd/lxd-site-mgrd --state-dir state_dir_2 &
	go run ./cmd/lxd-site-mgrd --state-dir state_dir_3 &

# Initialise cluster in development mode.
.PHONY: init-cluster-dev
init-cluster-dev:
	go run ./cmd/lxd-site-mgr --state-dir state_dir_1 init "member1" 127.0.0.1:9001 --bootstrap
	token_member2=$(go run ./cmd/lxd-site-mgr --state-dir state_dir_1 tokens add "member2")
	token_member3=$(go run ./cmd/lxd-site-mgr --state-dir state_dir_1 tokens add "member3")
	go run ./cmd/lxd-site-mgr --state-dir state_dir_2 init "member2" 127.0.0.1:9002 --token ${token_member2}
	go run ./cmd/lxd-site-mgr --state-dir state_dir_3 init "member3" 127.0.0.1:9003 --token ${token_member3}

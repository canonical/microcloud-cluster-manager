CGO_ENABLED?=0 # create statically linked binary
GOOS?=linux
GO_BIN?=app # name of the output application binary
GO?=go # name of the go binary
GOFLAGS?=-ldflags=-w -ldflags=-s -a # remove debug info, strip symbol table, force packages rebuild
GO_UI_FOLDER?=internal/app/management-api/api/v1/static
MAKEFLAGS += --no-print-directory
SHELL := /bin/bash # use bash shell

# export all variables defined as environment variables
.EXPORT_ALL_VARIABLES:

.PHONY: default
default: all

# ==============================================================================
# Static code linting utility targets.

.PHONY: add-hooks
add-hooks:
	@echo "Adding git hooks..."
	@cp scripts/pre-commit.sh .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Hooks Added"

.PHONY: lint-backend
lint-backend:
ifeq ($(shell command -v golangci-lint 2> /dev/null),)
	curl -sSfL https://golangci-lint.run/install.sh  | sh -s -- -b $$(go env GOPATH)/bin
endif
	golangci-lint run --timeout 10m

.PHONY: lint-ui-scss
lint-ui-scss:
	cd ui && yarn run lint-scss

.PHONY: lint-ui-js
lint-ui-js:
	cd ui && yarn run lint-js

# ==============================================================================
# Go module utility targets.

.PHONY: update-gomod
update-gomod:
	go get -t -v -u ./...
	go mod tidy

.PHONY: tidy-gomod
tidy-gomod:
	go mod tidy

# ====================================================================
# Local dev utility targets.

.PHONY: clean
clean:
	rm -rf cmd/app-coverage
	rm -rf internal/app/management-api/api/v1/static/ui
	rm -rf test/coverage
	rm -rf test/keys
	rm -rf ui/blog-report
	rm -rf ui/build
	rm -rf ui/coverage
	rm -rf ui/keys
	rm -rf ui/node_modules
	rm -rf ui/playwright-report
	rm -rf ui/test-results
	rm -rf ui/.dotrun.json
	rm -rf ui/haproxy-local.cfg
	rm -rf vendor
	rm -rf .cover

.PHONY: dev
dev:
	./scripts/run-backend.sh

# ====================================================================
# UI utilities
.PHONY: ui
ui:
	cd ui && dotrun

# ====================================================================
# test utilities

# make a request to the server to ensure it is up before running tests
.PHONY: ensure-service-running
ensure-service-running:
	@{ curl --insecure https://localhost:9000 > /dev/null 2>&1 || true; } 2>/dev/null

.PHONY: test-e2e
test-e2e:
	$(MAKE) ensure-service-running
	export management_api_cert_secret="$(shell pwd)/test/keys" ; \
	export cluster_connector_cert_secret="$(shell pwd)/test/keys" ; \
	export CLUSTER_CONNECTOR_PORT=9000 ; \
	go test -count=1 -v ./test/e2e

.PHONY: test-openapi
test-openapi:
	$(MAKE) ensure-service-running
	curl -Lo test/openapi/cluster-manager-api.yaml https://raw.githubusercontent.com/canonical/microcloud/main/doc/_extra/cluster-manager-api.yaml ; \
	export management_api_cert_secret="$(shell pwd)/test/keys" ; \
	export cluster_connector_cert_secret="$(shell pwd)/test/keys" ; \
	export CLUSTER_CONNECTOR_PORT=9000 ; \
	go test -count=1 -v ./test/openapi

.PHONY: test-ui-e2e
test-ui-e2e:
	cd ui && CI=$(CI) npx playwright test $(if $(PROJECT),--project $(PROJECT))

.PHONY: test-unit
test-unit:
	go test -count=1 -v ./test/unit

# ====================================================================
# CI build utilities for rockcraft

.PHONY: rock-version
rock-version:
	@awk -F': ' '/^version:/ {print $$2; exit} END {if (NR == 0) exit 1}' rockcraft.yaml | tr -d '"' || echo "Error: version not found in rockcraft.yaml"

.PHONY: rock-name
rock-name:
	@echo "microcloud-cluster-manager_$(shell $(MAKE) rock-version)_amd64.rock"

.PHONY: docker-image-name
docker-image-name:
	@echo "microcloud-cluster-manager:$(shell $(MAKE) rock-version)"

.PHONY: rock-to-docker
rock-to-docker:
	rockcraft.skopeo --insecure-policy copy \
		oci-archive:$(shell $(MAKE) rock-name) \
		docker-daemon:$(shell $(MAKE) docker-image-name)

# Output a docker image into tarball format, which can be side loaded into a microk8s cluster
# https://microk8s.io/docs/registry-images
.PHONY: docker-image-to-tarball
docker-image-to-tarball:
	docker save $(shell $(MAKE) docker-image-name) > microcloud-cluster-manager.tar

.PHONY: build-ui
build-ui:
	cd ui && yarn install --frozen=lockfile
	rm -rf ui/build
	cd ui && yarn build

.PHONY: copy-ui
copy-ui:
	rm -rf $(GO_UI_FOLDER)/ui
	mkdir -p $(GO_UI_FOLDER)
	cp -r ui/build/ui $(GO_UI_FOLDER)

# create a binary "app" located in project root
.PHONY: build
build: build-ui copy-ui
	$(GO) build -C cmd -o $(GO_BIN) ./

.PHONY: build-coverage
build-coverage: build-ui copy-ui
	$(GO) build -C cmd -cover -o app-coverage ./

# ====================================================================
# Development dependencies

.PHONY: install-core
install-core: install-go install-docker

.PHONY: install-deps
install-deps: install-nvm install-dotrun

# install golang based on version in go.mod if it does not exist
.PHONY: install-go
install-go:
	@if ! command -v go >/dev/null 2>&1; then \
		if [ -f "go.mod" ]; then \
			GO_VERSION=$$(grep -m1 "^go " go.mod | awk '{print $$2}'); \
			echo "\n---------> Installing Go version $$GO_VERSION from go.mod..."; \
			curl -OL https://go.dev/dl/go$${GO_VERSION}.linux-amd64.tar.gz && \
			sudo tar -C /usr/local -xzf go$${GO_VERSION}.linux-amd64.tar.gz && \
			rm go$${GO_VERSION}.linux-amd64.tar.gz; \
			if ! grep -q "/usr/local/go/bin" $$HOME/.bashrc; then \
				echo "Adding Go to PATH..."; \
				echo 'export PATH=$$PATH:/usr/local/go/bin' >> $$HOME/.bashrc; \
				source $$HOME/.bashrc; \
			fi; \
		else \
			echo "No go.mod found and Go not installed. Please specify a Go version."; \
			exit 1; \
		fi; \
	else \
		echo "Go is already installed."; \
	fi

# install docker if it does not exist
.PHONY: install-docker
install-docker:
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "\n---------> Installing Docker..."; \
		sudo snap install docker; \
		echo "Configuring Docker for user $$USER..."; \
		sudo addgroup --system docker; \
		sudo adduser $$USER docker; \
		sudo snap disable docker; \
		sudo snap enable docker; \
		if ! grep -q "/snap/bin" $$HOME/.bashrc; then \
			echo "Adding /snap/bin to PATH..."; \
			echo 'export PATH=$$PATH:/snap/bin' >> $$HOME/.bashrc; \
			source $$HOME/.bashrc; \
		fi; \
		echo "Restarting shell to apply Docker group changes..."; \
		echo "Please restart make install-deps to continue installation."; \
		newgrp docker; \
	else \
		echo "Docker is already installed."; \
	fi

# install nvm if it does not exist
.PHONY: install-nvm
install-nvm:
	@if [ ! -d "$$HOME/.nvm" ]; then \
		echo "\n---------> Installing NVM..."; \
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash; \
		export NVM_DIR="$$HOME/.nvm"; \
		[ -s "$$NVM_DIR/nvm.sh" ] && . "$$NVM_DIR/nvm.sh"; \
		nvm install 22; \
		nvm use 22; \
	else \
		echo "NVM is already installed."; \
	fi

# install dotrun if it does not exist
.PHONY: install-dotrun
install-dotrun:
	@if ! command -v dotrun >/dev/null 2>&1; then \
		echo "\n---------> Installing dotrun..."; \
		curl -sSL https://raw.githubusercontent.com/canonical/dotrun/main/scripts/install.sh | bash; \
		pipx ensurepath; \
		source $$HOME/.bashrc; \
	else \
		echo "dotrun is already installed."; \
	fi

# add local host entries to /etc/hosts
.PHONY: add-hosts
add-hosts:
	@echo "Adding microcloud-cluster-manager local entries to /etc/hosts..."
	@if ! grep -q "# microcloud-cluster-manager local" /etc/hosts; then \
		sudo sh -c 'echo "# microcloud-cluster-manager local" >> /etc/hosts'; \
		sudo sh -c 'echo "127.0.0.1 ma.lxd-cm.local" >> /etc/hosts'; \
		sudo sh -c 'echo "127.0.0.1 cc.lxd-cm.local" >> /etc/hosts'; \
		echo "Entries added to /etc/hosts."; \
	else \
		echo "Entries already exist in /etc/hosts."; \
	fi

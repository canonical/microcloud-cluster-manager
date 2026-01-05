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

.PHONY: lint-backend
lint-backend:
ifeq ($(shell command -v golangci-lint 2> /dev/null),)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
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
# Local dev cluster utility targets. (k8s, kustomize, kind, skaffold)

KIND_CLUSTER := dev-cluster

.PHONY: start-cluster
start-cluster:
	@if ! kind get clusters | grep -q "$(KIND_CLUSTER)"; then \
		echo "Cluster '$(KIND_CLUSTER)' does not exist. Creating..."; \
		kind create cluster \
			--image kindest/node:v1.31.0 \
			--name $(KIND_CLUSTER) \
			--config deployment/k8s/kind/kind-config.yaml; \
		kubectl config set-context --current --namespace=default; \
	else \
		echo "Cluster '$(KIND_CLUSTER)' already exists."; \
	fi

.PHONY: delete-cluster
delete-cluster:
	kind delete cluster --name $(KIND_CLUSTER)

.PHONY: dev-k8s-deploy
dev-k8s-deploy:
	# create custom tmp directory for skaffold to store build artifacts
	@if [ ! -d "./tmp" ]; then \
		mkdir tmp; \
	fi
	TMPDIR=tmp skaffold dev --no-prune=false -p docker

.PHONY: debug-k8s-deploy
debug-k8s-deploy:
	@if [ ! -d "./tmp" ]; then \
		mkdir tmp; \
	fi
	TMPDIR=tmp skaffold dev --no-prune=false -p debug

.PHONY: rock-k8s-deploy
rock-k8s-deploy:
	@if [ ! -d "./tmp" ]; then \
		mkdir tmp; \
	fi
	TMPDIR=tmp skaffold dev --no-prune=false --cache-artifacts=false -p rock

# unfortunately necessary as skaffold does not automatically remove images after removing k8s cluster objects
.PHONY: clean-dev
clean-dev:
	rm -rf tmp
	docker container prune -f --filter "label=io.x-k8s.kind.cluster=dev-cluster"
	docker images -f "dangling=true" -q | xargs -r docker rmi
	docker images --filter=reference='microcloud-cluster-manager:*' -q | xargs -I {} docker rmi {} -f

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
	./test/run-backend.sh

.PHONY: dev-rock
dev-rock: start-cluster dev-juju-setup rock-k8s-deploy

.PHONY: nuke
nuke: clean-dev delete-cluster dev-clean-juju

# ====================================================================
# UI utilities
.PHONY: ui
ui:
	cd ui && dotrun

# ====================================================================
# test utilities

# to ensure that all pods are ready before running tests, we check the liveliness of the pods
# rollout restart seems to break k8s portforwarding, here we make a request to the server to ensure it is up as well as reset the portforwarding
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

.PHONY: test-ui-e2e
test-ui-e2e:
	cd ui && CI=$(CI) npx playwright test $(if $(PROJECT),--project $(PROJECT))

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
	rm -rf $(GO_UI_FOLDER)
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
# CI k8s deployment utilities

.PHONY: deploy-cert-manager
deploy-cert-manager:
	@echo "Installing cert-manager.."
	kubectl apply -f deployment/k8s/cicd/cert/cert-manager.yaml
	@echo "Waiting for Cert-Manager deployment to become available..."
	kubectl wait --for=condition=available --timeout=300s deployment --all -n cert-manager
	@echo "Applying ClusterIssuer..."
	kubectl apply -f deployment/k8s/cicd/cert/cert-issuer.yaml
	@echo "Applying Certificates..."
	kubectl apply -f deployment/k8s/cicd/cert/management-api-cert.yaml
	kubectl apply -f deployment/k8s/cicd/cert/cluster-connector-cert.yaml
	@echo "Waiting for the certificate Secrets to be created..."
	kubectl wait --for=create --timeout=600s secret/management-api-cert-secret -n default
	kubectl wait --for=create --timeout=600s secret/cluster-connector-cert-secret -n default
	@echo "Certificates are ready!"

.PHONY: deploy-db
deploy-db:
	@echo "Deploying Postgres database..."
	kubectl apply -f deployment/k8s/cicd/db/config.yaml
	kubectl apply -f deployment/k8s/cicd/db/pv.yaml
	kubectl apply -f deployment/k8s/cicd/db/pvc.yaml
	kubectl apply -f deployment/k8s/cicd/db/svc.yaml
	kubectl apply -f deployment/k8s/cicd/db/ss.yaml
	kubectl rollout status --watch --timeout=600s statefulset/db-ss
	@echo "Postgres database is ready!"

.PHONY: deploy-configs
deploy-configs:
	@echo "Deploying configs..."
	sed -i 's/PROMETHEUS_ADDRESS/$(PROMETHEUS_ADDRESS)/g' deployment/k8s/cicd/config/config.yaml
	kubectl apply -f deployment/k8s/cicd/config/config.yaml
	kubectl wait --for=create --timeout=600s cm/config -n default
	@echo "Configs is ready!"

.PHONY: deploy-management-api
deploy-management-api:
	@echo "Deploying management-api..."
	sed -i 's/IMAGE_NAME/$(IMAGE_NAME)/g' deployment/k8s/cicd/management-api/depl.yaml
	kubectl apply -f deployment/k8s/cicd/management-api/svc.yaml
	kubectl apply -f deployment/k8s/cicd/management-api/depl.yaml
	kubectl rollout status --watch --timeout=600s deployment/management-api-depl
	@echo "Management-api is ready!"

.PHONY: deploy-cluster-connector
deploy-cluster-connector:
	@echo "Deploying cluster-connector..."
	sed -i 's/IMAGE_NAME/$(IMAGE_NAME)/g' deployment/k8s/cicd/cluster-connector/depl.yaml
	kubectl apply -f deployment/k8s/cicd/cluster-connector/svc.yaml
	kubectl apply -f deployment/k8s/cicd/cluster-connector/depl.yaml
	kubectl rollout status --watch --timeout=600s deployment/cluster-connector-depl
	@echo "Cluster-connector is ready!"

.PHONY: deploy-ingress
deploy-ingress:
	@echo "Deploying HAProxy ingress..."
	kubectl apply -f deployment/k8s/cicd/ingress/ingress.yaml
	kubectl apply -f deployment/k8s/cicd/ingress/haproxy.yaml
	kubectl rollout status --watch --timeout=600s -n haproxy-controller deployment/haproxy-kubernetes-ingress
	@echo "Ingress is ready!"

.PHONY: deploy-ci-k8s-cluster
deploy-ci-k8s-cluster:
	$(MAKE) deploy-cert-manager
	$(MAKE) deploy-db
	$(MAKE) deploy-configs PROMETHEUS_ADDRESS=$(PROMETHEUS_ADDRESS)
	$(MAKE) deploy-management-api IMAGE_NAME=$(IMAGE_NAME)
	$(MAKE) deploy-cluster-connector IMAGE_NAME=$(IMAGE_NAME)
	$(MAKE) deploy-ingress
	$(MAKE) add-hosts

# ====================================================================
# Development dependencies

.PHONY: install-deps
install-deps: install-go install-docker install-kubectl \
							install-kind install-skaffold install-nvm \
							install-juju install-dotrun

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
		newgrp docker; \
		sudo snap disable docker; \
		sudo snap enable docker; \
		if ! grep -q "/snap/bin" $$HOME/.bashrc; then \
			echo "Adding /snap/bin to PATH..."; \
			echo 'export PATH=$$PATH:/snap/bin' >> $$HOME/.bashrc; \
			source $$HOME/.bashrc; \
		fi; \
	else \
		echo "Docker is already installed."; \
	fi

# install kubernetes controller if it does not exist
.PHONY: install-kubectl
install-kubectl:
	@if ! command -v kubectl >/dev/null 2>&1; then \
		echo "\n---------> Installing kubectl..."; \
		KUBECTL_VERSION=$$(curl -L -s https://dl.k8s.io/release/stable.txt); \
		curl -LO "https://dl.k8s.io/release/$${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && \
		sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
		rm kubectl; \
	else \
		echo "kubectl is already installed."; \
	fi

# install kind if it does not exist
.PHONY: install-kind
install-kind:
	@if ! command -v kind >/dev/null 2>&1; then \
		echo "\n---------> Installing Kind..."; \
		curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.24.0/kind-linux-amd64; \
		chmod +x ./kind; \
		sudo mv ./kind /usr/local/bin/kind; \
	else \
		echo "Kind is already installed."; \
	fi

# install skaffold if it does not exist
.PHONY: install-skaffold
install-skaffold:
	@if ! command -v skaffold >/dev/null 2>&1; then \
		echo "\n---------> Installing Skaffold..."; \
		curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && \
		chmod +x skaffold && sudo mv skaffold /usr/local/bin; \
	else \
		echo "Skaffold is already installed."; \
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

# install juju if it does not exist
.PHONY: install-juju
install-juju:
	@if ! command -v juju >/dev/null 2>&1; then \
		echo "\n---------> Installing Juju..."; \
		sudo snap install juju --channel=3.6/stable; \
	else \
		echo "Juju is already installed."; \
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

# ====================================================================
# juju setup utilities

# development juju setup
.PHONY: dev-juju-setup
dev-juju-setup:
	@echo "Setting up Juju controller for development..."
	@if ! juju clouds --client --all | grep cluster-manager; then \
		juju add-k8s --context-name kind-dev-cluster cluster-manager; \
		juju bootstrap cluster-manager cm-controller; \
	else \
		echo "Juju controller already exists."; \
	fi

.PHONY: dev-cos-deploy
dev-cos-deploy:
	@echo "Deploying COS-Lite to the dev cluster..."
	@if ! kubectl get ns | grep cos; then \
		echo "Applying DNS configuration..."; \
		kubectl apply -f deployment/k8s/dev/dns; \
		@echo "Installing MetalLB for COS-Lite ingress..."; \
		kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.14.9/config/manifests/metallb-native.yaml; \
		kubectl wait --for=condition=available --timeout=300s deployment --all -n metallb-system; \
		kubectl rollout status daemonset speaker -n metallb-system --timeout 120s; \
		echo "Setting MetalLB address pool..."; \
		ip_addr=$$(ip -4 -j route get 2.2.2.2 | jq -r '.[] | .prefsrc'); \
		ip_range=$${ip_addr}-$${ip_addr}; \
		ip_range_str="addresses:\\n      - $${ip_range}"; \
		sed "s@{{addresses}}@$${ip_range_str}@g" deployment/k8s/dev/metallb/addresspool.yaml | kubectl apply -f -; \
		echo "Deploying COS-Lite to the dev cluster..."; \
		juju add-model cos && juju switch cos; \
		juju deploy cos-lite --trust; \
	else \
		echo "COS-Lite already deployed."; \
	fi

.PHONY: dev-clean-juju
dev-clean-juju:
	juju unregister cm-controller --no-prompt
	juju remove-cloud cluster-manager

.PHONY: update-traefik-port
update-traefik-port:
	@SERVICE_NAME="traefik-lb"; \
	NAMESPACE="cos"; \
	TIMEOUT=120; \
	START_TIME=$$(date +%s); \
	echo "Waiting for service $$SERVICE_NAME to exist in namespace $$NAMESPACE..."; \
	while true; do \
	    if kubectl get service "$$SERVICE_NAME" -n "$$NAMESPACE" > /dev/null 2>&1; then \
	        echo "Service $$SERVICE_NAME exists!"; \
	        break; \
	    fi; \
	    CURRENT_TIME=$$(date +%s); \
	    ELAPSED_TIME=$$((CURRENT_TIME - START_TIME)); \
	    if [ "$$ELAPSED_TIME" -ge "$$TIMEOUT" ]; then \
	        echo "Timed out waiting for service $$SERVICE_NAME to exist."; \
	        exit 1; \
	    fi; \
	    sleep 2; \
	done
	kubectl patch svc/traefik-lb -n cos -p '{"spec":{"ports":[{"port":80, "nodePort":30001}]}}'

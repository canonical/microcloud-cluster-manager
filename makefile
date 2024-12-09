GOMIN=1.23.4
CGO_ENABLED?=0 # create statically linked binary
GOOS?=linux
GO_BIN?=app # name of the output application binary
GO?=go # name of the go binary
GOFLAGS?=-ldflags=-w -ldflags=-s -a # remove debug info, strip symbol table, force packages rebuild
GO_UI_FOLDER?=internal/app/management-api/api/v1/static
MAKEFLAGS += --no-print-directory

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
	go get -t -v -d -u ./...
	go mod tidy -go=$(GOMIN)

.PHONY: tidy-gomod
tidy-gomod:
	go mod tidy -go=$(GOMIN)

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
	skaffold dev --no-prune=false -p docker

.PHONY: debug-k8s-deploy
debug-k8s-deploy:
	skaffold dev --no-prune=false -p debug

.PHONY: rock-k8s-deploy
rock-k8s-deploy:
	skaffold dev --no-prune=false --cache-artifacts=false -p rock

# unfortunately necessary as skaffold does not automatically remove images after removing k8s cluster objects
.PHONY: clean-dev
clean-dev:
	docker container prune -f
	docker images -f "dangling=true" -q | xargs -r docker rmi
	docker images --filter=reference='lxd-cluster-manager:*' -q | xargs -I {} docker rmi {} -f

.PHONY: dev
dev: start-cluster dev-k8s-deploy

.PHONY: debug
debug: start-cluster debug-k8s-deploy

.PHONY: dev-rock
dev-rock: start-cluster rock-k8s-deploy

.PHONY: nuke
nuke: clean-dev delete-cluster

# ====================================================================
# UI utilities
.PHONY: ui
ui: 
	cd ui && dotrun

# ====================================================================
# test utilities

# to ensure that all pods are ready before running tests, we check the liveliness of the pods
# rollout restart seems to break k8s portforwarding, here we make a request to the server to ensure it is up as well as reset the portforwarding
.PHONY: switch-test-mode
switch-test-mode:
	kubectl patch configmap config --patch '{"data":{"TEST_MODE":"$(IS_ON)"}}'
	kubectl rollout restart deployment/management-api-depl
	kubectl rollout status deployment/management-api-depl --timeout=300s
	@{ curl --insecure https://localhost:9000 > /dev/null 2>&1 || true; } 2>/dev/null

# Need to set TEST_MODE to true in the management-api deployment so we can by pass oidc authentication
.PHONY: test-e2e
test-e2e: 
	$(MAKE) switch-test-mode IS_ON=true
	go test -count=1 -v ./test/e2e
	$(MAKE) switch-test-mode IS_ON=false

.PHONY: test-ui-e2e
test-ui-e2e:
	cd ui && CI=$(CI) npx playwright test --project $(PROJECT)

# ====================================================================
# CI build utilities for rockcraft

.PHONY: rock-version
rock-version:
	@awk -F': ' '/^version:/ {print $$2; exit} END {if (NR == 0) exit 1}' rockcraft.yaml | tr -d '"' || echo "Error: version not found in rockcraft.yaml"

.PHONY: rock-name
rock-name:
	@echo "lxd-cluster-manager_$(shell $(MAKE) rock-version)_amd64.rock"

.PHONY: docker-image-name
docker-image-name:
	@echo "lxd-cluster-manager:$(shell $(MAKE) rock-version)"

.PHONY: rock-to-docker
rock-to-docker:
	rockcraft.skopeo --insecure-policy copy \
		oci-archive:$(shell $(MAKE) rock-name) \
		docker-daemon:$(shell $(MAKE) docker-image-name)

# Output a docker image into tarball format, which can be side loaded into a microk8s cluster
# https://microk8s.io/docs/registry-images
.PHONY: docker-image-to-tarball
docker-image-to-tarball:
	docker save $(shell $(MAKE) docker-image-name) > lxd-cluster-manager.tar

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

.PHONY: expose-services
expose-services:
	@echo "Exposing management-api and cluster-connector..."
	@( \
		services="svc/management-api-svc:9000:management-api svc/cluster-connector-svc:9001:cluster-conn"; \
		while true; do \
			for svc in $$services; do \
				svc_name=$$(echo $$svc | cut -d':' -f1); \
				local_port=$$(echo $$svc | cut -d':' -f2); \
				target_port=$$(echo $$svc | cut -d':' -f3); \
				\
				# Check if port-forwarding is already active \
				if ! lsof -i :$$local_port > /dev/null; then \
					echo "Reconnecting to $$svc_name..."; \
					kubectl port-forward $$svc_name $$local_port:$$target_port & \
				fi; \
			done; \
			sleep 5; \
		done; \
	) &
	@echo "management-api and cluster-connector are exposed on localhost:9000 and localhost:9001 respectively"

.PHONY: deploy-ci-k8s-cluster
deploy-ci-k8s-cluster:
	$(MAKE) deploy-cert-manager
	$(MAKE) deploy-db
	$(MAKE) deploy-configs
	$(MAKE) deploy-management-api IMAGE_NAME=$(IMAGE_NAME)
	$(MAKE) deploy-cluster-connector IMAGE_NAME=$(IMAGE_NAME)
	$(MAKE) expose-services

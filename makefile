GOMIN=1.22.7

.PHONY: default
default: all

# ==============================================================================
# Static code linting utility targets.

.PHONY: check
check: check-static check-unit check-system

.PHONY: check-unit
check-unit:
ifeq "$(GOCOVERDIR)" ""
	go test ./...
else
	go test ./... -cover -test.gocoverdir="${GOCOVERDIR}"
endif

.PHONY: check-system
check-system:
	true

.PHONY: check-static
check-static:
ifeq ($(shell command -v golangci-lint 2> /dev/null),)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
endif
	golangci-lint run --timeout 5m

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

# Check status for all resources in the cluster (across all name spaces)
.PHONY: cluster-status
cluster-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

.PHONY: dev-k8s-configs
dev-k8s-configs:
	kubectl kustomize deployment/k8s/dev

.PHONY: dev-k8s-deploy
dev-k8s-deploy:
	skaffold dev --no-prune=false

.PHONY: debug-k8s-deploy
debug-k8s-deploy:
	skaffold dev --no-prune=false -p debug

# unfortunately necessary as skaffold does not automatically remove images after removing k8s cluster objects
.PHONY: clean-dev
clean-dev:
	skaffold delete
	docker container prune -f
	docker images -f "dangling=true" -q | xargs -r docker rmi
	docker images --filter=reference='cluster-manager-img:*' -q | xargs -I {} docker rmi {} -f

.PHONY: dev
dev: start-cluster dev-k8s-deploy

.PHONY: debug
debug: start-cluster debug-k8s-deploy

.PHONY: nuke
nuke: clean-dev delete-cluster

# ====================================================================
# UI utilities
.PHONY: ui
ui: 
	cd ui && dotrun

# ====================================================================
# dev database utilities

.PHONY: migrate-db
migrate-db:
	go run cmd/admin/main.go

# ====================================================================
# test utilities

# management API tests
# POST /1.0/remote-cluster-join-token
# curl -X POST http://localhost:8414/1.0/remote-cluster-join-token -H "Content-Type: application/json" -d '{ "expiry": "2024-11-21T15:30:00Z", "cluster_name": "example-cluster" }'
# GET /1.0/remote-cluster-join-token
# curl -X GET http://localhost:8414/1.0/remote-cluster-join-token -H "Content-Type: application/json"
# DELETE /1.0/remote-cluster-join-token/:name
# curl -X DELETE http://localhost:8414/1.0/remote-cluster-join-token/example-cluster -H "Content-Type: application/json"
# GET /1.0/remote-cluster
# curl -X GET http://localhost:8414/1.0/remote-cluster -H "Content-Type: application/json"
# GET /1.0/remote-cluster/:name
# curl -X GET http://localhost:8414/1.0/remote-cluster/test -H "Content-Type: application/json"
# PATCH /1.0/remote-cluster-join-token/:name
# curl -X PATCH http://localhost:8414/1.0/remote-cluster/test -H "Content-Type: application/json" -d '{ "status": "ACTIVE" }'
# DELETE /1.0/remote-cluster/:name
# curl -X DELETE http://localhost:8414/1.0/remote-cluster/test -H "Content-Type: application/json"

# control API tests
# POST /1.0/remote-cluster
# curl -X POST http://localhost:8415/1.0/remote-cluster -H "Content-Type: application/json" -d '{ "cluster_certificate": "abc", "cluster_name": "test-cluster" }'
# POST /1.0/remote-cluster/status
# curl -X POST http://localhost:8415/1.0/remote-cluster/status -H "Content-Type: application/json" -d '{ "cpu_total_count": 20, "cpu_load_1": "0.1", "cpu_load_5": "0.2", "cpu_load_15": "0.3", "memory_total_amount": 30, "memory_usage": 40, "disk_total_size": 50, "disk_usage": 60, "member_statuses": [{"status": "RUNNING", "count": 10}, {"status": "DOWN", "count": 100}], "instance_statuses": [{"status": "RUNNING", "count": 10}, {"status": "DOWN", "count": 100}], "cluster_name": "test-cluster" }'
# DELETE /1.0/remote-cluster/:name
# curl -X DELETE http://localhost:8415/1.0/remote-cluster/test-cluster -H "Content-Type: application/json"
.PHONY: test-e2e
test-e2e:
	go test -count=1 -v ./test/e2e

.PHONY: test-ui-e2e
test-ui-e2e:
	cd ui && npx playwright test
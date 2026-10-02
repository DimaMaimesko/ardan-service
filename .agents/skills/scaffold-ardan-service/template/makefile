# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# Deploy First Mentality
# ==============================================================================
# Install Tooling and Dependencies
#
#	This project uses Docker and it is expected to be installed. Please provide
#	Docker at least 4 CPUs.
#
#	Run these commands to install everything needed.
#	$ make dev-brew
#	$ make dev-docker
#	$ make dev-gotooling
#
# ==============================================================================
# Running Tests
#
#	The tests start a Postgres container named tmplsvctest and create a fresh
#	database per test.
#
#	$ make test
#	$ make test-down
#
# ==============================================================================
# Running The Project
#
#	Kind (Kubernetes) with Tempo and Grafana:
#	$ make dev-run
#	$ make token
#	$ export TOKEN=<COPY TOKEN>
#	$ make users
#	$ make grafana
#
#	Docker Compose (no tracing):
#	$ make compose-build-up
#
#	Local binary against a local database:
#	$ make run
#
# ==============================================================================
# Dev Keys
#
#	zarf/keys/<kid>.pem is a development signing key generated for this
#	project. Never use it outside development. Production keys come from a
#	secret: mounted files at /service/zarf/keys or TMPLSVC_AUTH_KEYS_JSON.

# ==============================================================================
# Define dependencies

GOLANG          := golang:1.27
ALPINE          := alpine:3.24
KIND            := kindest/node:v1.37.0
POSTGRES        := postgres:18.6
GRAFANA         := grafana/grafana:13.2.1
TEMPO           := grafana/tempo:3.0.3

KIND_CLUSTER    := tmplsvc-cluster
NAMESPACE       := tmplsvc-system
APP             := tmplsvc
VERSION         := 0.0.1
SERVICE_IMAGE   := localhost/$(APP):$(VERSION)
KID             := TMPLKID
CHART           := zarf/helm/charts/$(APP)

# ==============================================================================
# Install dependencies

dev-gotooling:
	go install github.com/divan/expvarmon@latest
	go install github.com/rakyll/hey@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest

dev-brew:
	brew update
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list helm || brew install helm
	brew list pgcli || brew install pgcli
	brew list watch || brew install watch

dev-docker:
	docker pull $(GOLANG) & \
	docker pull $(ALPINE) & \
	docker pull $(KIND) & \
	docker pull $(POSTGRES) & \
	docker pull $(GRAFANA) & \
	docker pull $(TEMPO) & \
	wait;

# ==============================================================================
# Building containers

build: service

service:
	docker build \
		-f zarf/docker/dockerfile.$(APP) \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_TAG=$(VERSION) \
		--build-arg BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		.

# ==============================================================================
# Running from within k8s/kind
#
# Images are copied into the Kind node with docker save | ctr import. This
# avoids the "content digest not found" failures some Docker versions have
# with kind load docker-image, and it works without a registry or network.

dev-up:
	kind create cluster \
		--image $(KIND) \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml

	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	docker save $(POSTGRES) | docker exec -i $(KIND_CLUSTER)-control-plane ctr --namespace=k8s.io images import - & \
	docker save $(GRAFANA) | docker exec -i $(KIND_CLUSTER)-control-plane ctr --namespace=k8s.io images import - & \
	docker save $(TEMPO) | docker exec -i $(KIND_CLUSTER)-control-plane ctr --namespace=k8s.io images import - & \
	wait;

dev-down:
	kind delete cluster --name $(KIND_CLUSTER)

dev-status-all:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-status:
	watch -n 2 kubectl get pods -o wide --all-namespaces

# ------------------------------------------------------------------------------

dev-load:
	docker save $(SERVICE_IMAGE) | docker exec -i $(KIND_CLUSTER)-control-plane ctr --namespace=k8s.io images import -

dev-apply:
	kubectl apply -k zarf/k8s/dev/tempo
	kubectl apply -k zarf/k8s/dev/grafana

	kubectl apply -k zarf/k8s/dev/database
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/database

	helm upgrade --install $(APP) $(CHART) \
		--namespace $(NAMESPACE) \
		--values $(CHART)/values-dev.yaml \
		--set auth.keys.kid=$(KID) \
		--set-file auth.keys.pem=zarf/keys/$(KID).pem \
		--wait \
		--timeout 5m

dev-restart:
	kubectl rollout restart deployment $(APP) --namespace=$(NAMESPACE)

dev-run: build dev-up dev-load dev-apply

dev-update: build dev-load dev-restart

dev-update-apply: build dev-load dev-apply

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 --max-log-requests=6 | go run api/tooling/logfmt/main.go -service=$(APP)

dev-logs-init:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) -f --tail=100 -c init-migrate-seed

dev-describe-deployment:
	kubectl describe deployment --namespace=$(NAMESPACE) $(APP)

dev-describe-pod:
	kubectl describe pod --namespace=$(NAMESPACE) -l app=$(APP)

dev-db-shell:
	kubectl exec --namespace=$(NAMESPACE) -it sts/database -- psql -U postgres

dev-database-restart:
	kubectl rollout restart statefulset database --namespace=$(NAMESPACE)

dev-uninstall:
	helm uninstall $(APP) --namespace $(NAMESPACE) --ignore-not-found

# ==============================================================================
# Helm

chart-lint:
	helm lint $(CHART)
	helm lint $(CHART) --values $(CHART)/values-dev.yaml

chart-template:
	helm template $(APP) $(CHART) --values $(CHART)/values-dev.yaml --debug

# ==============================================================================
# Docker Compose

compose-up:
	cd ./zarf/compose/ && docker compose -f docker_compose.yaml -p $(APP) up -d

compose-build-up: build compose-up

compose-down:
	cd ./zarf/compose/ && docker compose -f docker_compose.yaml -p $(APP) down

compose-logs:
	cd ./zarf/compose/ && docker compose -f docker_compose.yaml -p $(APP) logs

# ==============================================================================
# Running the binary locally against a database on localhost:5432

run:
	TMPLSVC_DB_HOST=localhost TMPLSVC_TEMPO_HOST= go run api/services/$(APP)/main.go | go run api/tooling/logfmt/main.go

run-help:
	go run api/services/$(APP)/main.go --help

# ==============================================================================
# Administration

migrate:
	TMPLSVC_DB_HOST=localhost go run api/tooling/admin/main.go migrate

seed: migrate
	TMPLSVC_DB_HOST=localhost go run api/tooling/admin/main.go seed

pgcli:
	pgcli postgresql://postgres:postgres@localhost

token-gen:
	TMPLSVC_DB_HOST=localhost go run api/tooling/admin/main.go gentoken 5cf37266-3473-4006-984f-9325122678b7 $(KID)

# ==============================================================================
# Hitting endpoints

liveness:
	curl -i http://localhost:3000/v1/liveness

readiness:
	curl -i http://localhost:3000/v1/readiness

token:
	curl -s --user "admin@example.com:gophers" http://localhost:3000/v1/auth/token/$(KID)

users:
	curl -i -H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/v1/users?page=1&rows=2"

products:
	curl -i -H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/v1/products?page=1&rows=2"

load:
	hey -m GET -c 100 -n 1000 -H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/v1/users?page=1&rows=2"

otel-test:
	curl -i \
	-H "Traceparent: 00-918dd5ecf264712262b68cf2ef8b5239-896d90f23f69f006-01" \
	--user "admin@example.com:gophers" http://localhost:3000/v1/auth/token/$(KID)

# ==============================================================================
# Metrics and Tracing

metrics-view:
	expvarmon -ports="localhost:3010" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

grafana:
	open -a "Google Chrome" http://localhost:3100/ || xdg-open http://localhost:3100/

statsviz:
	open -a "Google Chrome" http://localhost:3010/debug/statsviz || xdg-open http://localhost:3010/debug/statsviz

# ==============================================================================
# Running tests within the local computer

test-down:
	docker stop tmplsvctest
	docker rm tmplsvctest -v

test-only:
	CGO_ENABLED=0 go test -count=1 ./...

test-r:
	CGO_ENABLED=1 go test -race -count=1 ./...

fmt:
	go fmt ./...

lint:
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...

vuln-check:
	govulncheck ./...

test: test-only lint vuln-check

test-race: test-r lint vuln-check

# ==============================================================================
# Modules support

tidy:
	go mod tidy
	go mod vendor

deps-list:
	go list -m -u -mod=readonly all

deps-upgrade:
	go get -u -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

# ==============================================================================
# Help

help:
	@echo "Usage: make <command>"
	@echo ""
	@echo "Setup:      dev-brew dev-docker dev-gotooling"
	@echo "Test:       test test-only test-race test-down lint vuln-check fmt"
	@echo "Kind:       dev-run dev-up dev-down dev-load dev-apply dev-update dev-update-apply dev-logs dev-status"
	@echo "Helm:       chart-lint chart-template dev-uninstall"
	@echo "Compose:    compose-build-up compose-up compose-down compose-logs"
	@echo "Local:      run run-help migrate seed pgcli token-gen"
	@echo "Endpoints:  liveness readiness token users products load otel-test"
	@echo "Observe:    grafana statsviz metrics-view"
	@echo "Modules:    tidy deps-list deps-upgrade deps-cleancache"

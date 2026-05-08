## quantum-tiktok-operator Makefile
## Standard controller-runtime project layout.

# Image settings
IMG ?= ghcr.io/yourusername/quantum-tiktok-operator:latest
PLATFORMS ?= linux/arm64,linux/amd64

# Go settings
GOBIN ?= $(shell go env GOPATH)/bin
ENVTEST_K8S_VERSION = 1.29.0

# Tools
CONTROLLER_GEN = $(GOBIN)/controller-gen
ENVTEST         = $(GOBIN)/setup-envtest
GOLANGCI_LINT   = $(GOBIN)/golangci-lint

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate CRD and RBAC manifests.
	$(CONTROLLER_GEN) rbac:roleName=operator-role crd webhook paths="./..." output:crd:artifacts:config=config/crd output:rbac:artifacts:config=config/rbac

.PHONY: generate
generate: controller-gen ## Generate DeepCopy methods.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint.
	$(GOLANGCI_LINT) run

.PHONY: test
test: manifests generate envtest ## Run unit and integration tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(GOBIN) -p path)" \
		go test ./... -coverprofile cover.out

.PHONY: verify-coherence
verify-coherence: ## Verify cluster coherence (non-deterministic, ~60% pass rate).
	@echo "Sampling quantum state..."
	@bash hack/verify-coherence.sh

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build operator binary.
	go build -o bin/operator ./cmd/operator/

.PHONY: run
run: manifests generate fmt vet ## Run operator against current kubeconfig.
	go run ./cmd/operator/ \
		--metrics-addr=:8080 \
		--leader-elect=false

.PHONY: docker-build
docker-build: ## Build container image.
	docker buildx build --platform $(PLATFORMS) -t $(IMG) .

.PHONY: docker-push
docker-push: ## Push container image.
	docker push $(IMG)

##@ Deploy

.PHONY: install
install: manifests ## Install CRDs into the cluster.
	kubectl apply -f config/crd/

.PHONY: uninstall
uninstall: ## Remove CRDs from the cluster.
	kubectl delete -f config/crd/ --ignore-not-found

.PHONY: deploy
deploy: manifests ## Deploy operator to the cluster.
	kubectl apply -f config/rbac/
	kubectl apply -f config/deploy/

.PHONY: undeploy
undeploy: ## Remove operator from the cluster.
	kubectl delete -f config/deploy/ --ignore-not-found
	kubectl delete -f config/rbac/ --ignore-not-found

##@ Tools

.PHONY: controller-gen
controller-gen:
	test -f $(CONTROLLER_GEN) || go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

.PHONY: envtest
envtest:
	test -f $(ENVTEST) || go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: golangci-lint
golangci-lint:
	test -f $(GOLANGCI_LINT) || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

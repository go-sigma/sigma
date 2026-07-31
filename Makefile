GOCMD             = go
GOTEST            = $(GOCMD) test
GOVET             = $(GOCMD) vet
BINARY_NAME       = sigma
CLI_BINARY_NAME   = sigma-cli
VERSION          ?= $(shell git describe --tags --always)
SERVICE_PORT     ?= 3000
DOCKER_REGISTRY  ?= ghcr.io/go-sigma

APPNAME          ?= sigma
NAMESPACE        ?= sigma
KUBECONFIG       ?= ~/.kube/config
REPOSITORY       ?= ghcr.io/go-sigma/sigma
TAG              ?= nightly-alpine
MIGRATION_NAME   ?=
RANDOM_PASSWORD  := $(shell openssl rand -base64 6 | tr -d '/+' | tr '[:upper:]' '[:lower:]' | head -c 8)

SHELL            := /bin/bash

GREEN            := $(shell tput -Txterm setaf 2)
YELLOW           := $(shell tput -Txterm setaf 3)
WHITE            := $(shell tput -Txterm setaf 7)
CYAN             := $(shell tput -Txterm setaf 6)
RESET            := $(shell tput -Txterm sgr0)

GOLDFLAGS        += -X github.com/go-sigma/sigma/pkg/version.Version=$(shell git describe --tags --always)
GOLDFLAGS        += -X github.com/go-sigma/sigma/pkg/version.BuildDate=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOLDFLAGS        += -X github.com/go-sigma/sigma/pkg/version.GitHash=$(shell git rev-parse --short HEAD)
GOFLAGS           = -ldflags '-s -w $(GOLDFLAGS)' -trimpath

GOOS             ?= $(shell go env GOOS)
DOCKER_GOOS      ?= $(if $(filter darwin,$(GOOS)),linux,$(GOOS))
GOARCH           ?= $(shell go env GOARCH)
CC               ?=
CXX              ?=

DOCKER_PLATFORMS ?= $(DOCKER_GOOS)/$(GOARCH)
USE_MIRROR       ?= false

.PHONY: all
all: build

## Build:
.PHONY: build
build: ## Build sigma and put the output binary in ./bin
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 GO111MODULE=on CC="$(CC)" CXX="$(CXX)" $(GOCMD) build $(GOFLAGS) -tags "netgo,timetzdata,exclude_graphdriver_btrfs,containers_image_openpgp" -o bin/$(BINARY_NAME) -v .

.PHONY: build-cli
build-cli: ## Build sigma CLI and put the output binary in ./bin
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 $(GOCMD) build $(GOFLAGS) -o bin/$(CLI_BINARY_NAME) -v ./cmd/cli

.PHONY: clean
clean: ## Remove build output files
	$(RM) ./bin/sigma ./bin/sigma-cli

.PHONY: vendor
vendor: ## Tidy Go module dependencies
	@$(GOCMD) mod tidy

## Lint:
.PHONY: lint
lint: lint-go lint-dockerfile ## Run all available linters

.PHONY: lint-dockerfile
lint-dockerfile: ## Lint Dockerfiles
	@hadolint --ignore DL3006 $(shell find build -name "Dockerfile*")

.PHONY: lint-go
lint-go: ## Run golangci-lint
	@golangci-lint run --timeout=10m --build-tags "netgo,timetzdata,exclude_graphdriver_btrfs,containers_image_openpgp"

## Docker:
.PHONY: docker-build
docker-build: ## Build sigma Docker images
	docker buildx build --file ./build/Dockerfile --target sigma-server \
		--platform $(DOCKER_PLATFORMS) \
		--build-arg USE_MIRROR=$(USE_MIRROR) \
		--progress plain --provenance false --sbom false --load -t $(REPOSITORY):$(TAG) .
	docker buildx build --file ./build/Dockerfile --target sigma-builder \
		--platform $(DOCKER_PLATFORMS) \
		--build-arg USE_MIRROR=$(USE_MIRROR) \
		--progress plain --provenance false --sbom false --load -t $(REPOSITORY)-builder:$(TAG) .

## Misc:
.PHONY: migration-create
migration-create: ## Create a new migration file
	@migrate create -dir ./pkg/dal/migrations/mysql -seq -digits 4 -ext sql $(MIGRATION_NAME)

.PHONY: sql-format
sql-format: ## Format all SQL migration files
	@find ${PWD}/pkg/dal/migrations/mysql -type f -iname "*.sql" | xargs -n1 sql-formatter -l mysql --fix
	@find ${PWD}/pkg/dal/migrations/sqlite3 -type f -iname "*.sql" | xargs -n1 sql-formatter -l sqlite --fix
	@find ${PWD}/pkg/dal/migrations/turso -type f -iname "*.sql" | xargs -n1 sql-formatter -l sqlite --fix
	@find ${PWD}/pkg/dal/migrations/postgresql -type f -iname "*.sql" | xargs -n1 sql-formatter -l postgresql --fix

.PHONY: changelog
changelog: ## Generate changelog
	# brew install git-cliff
	@git-cliff --config cliff.toml --tag $(VERSION) --output CHANGELOG.md

.PHONY: gormgen
gormgen: ## Generate GORM models from the database schema
	@$(GOCMD) run ./pkg/dal/cmd/gen.go

.PHONY: swagen
swagen: ## Generate Swagger documentation from code comments
	@$(GOCMD) tool swag fmt
	@$(GOCMD) tool swag init --output tools/skill/sigma-api-operator --outputTypes yaml

.PHONY: addlicense
addlicense: ## Add license headers to source files
	@find pkg -type f -name "*.go" | grep -v "pkg/handlers/apidocs/docs.go" | xargs addlicense -l apache -y 2026 -c "sigma"
	@find cmd -type f -name "*.go" | xargs addlicense -l apache -y 2026 -c "sigma"
	@addlicense -l apache -y 2026 -c "sigma" main.go
	@addlicense -l apache -y 2026 -c "sigma" web/web.go
	@find web/src -type f -name "*.tsx" | xargs addlicense -l apache -y 2026 -c "sigma"
	@find web/src -type f -name "*.ts" | xargs addlicense -l apache -y 2026 -c "sigma"
	@find web/src -type f -name "*.css" | xargs addlicense -l apache -y 2026 -c "sigma"

## Kube:
.PHONY: kube_install
kube_install: ## Install sigma on Kubernetes using Helm
	@if [ -z $(KUBECONFIG) ]; then \
		KUBECONFIG=$$HOME/.kube/config; \
	fi;
	@helm upgrade --install $(APPNAME) ./deploy/sigma --create-namespace --namespace $(NAMESPACE) \
		--set image.repository=$(REPOSITORY) \
		--set image.tag=$(TAG) \
		--set mysql.auth.rootPassword=$(RANDOM_PASSWORD) \
		--set mysql.auth.password=$(RANDOM_PASSWORD) \
		--set redis.auth.password=$(RANDOM_PASSWORD) \
		--set minio.secretKey=$(RANDOM_PASSWORD) \
		--kubeconfig $(KUBECONFIG)

.PHONY: kube_uninstall
kube_uninstall: ## Uninstall sigma from Kubernetes using Helm
	@KUBECONFIG=$(KUBECONFIG)
	@helm uninstall $(APPNAME) -n$(NAMESPACE)

## Help:
.PHONY: help
help: ## Show this help
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-30s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

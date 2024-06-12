GOCMD             = go
GOTEST            = $(GOCMD) test
GOVET             = $(GOCMD) vet
BINARY_NAME       = sigma
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

GOOS             ?= linux
GOARCH           ?= arm64
CC               ?=
CXX              ?=

DOCKER_PLATFORMS ?= $(GOOS)/$(GOARCH)
USE_MIRROR       ?= false
WITH_TRIVY_DB    ?= false

.PHONY: all test build vendor

all: build build-builder

## Build:
build: ## Build sigma and put the output binary in ./bin
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 GO111MODULE=on CC="$(CC)" CXX="$(CXX)" $(GOCMD) build $(GOFLAGS) -tags "timetzdata,exclude_graphdriver_devicemapper,exclude_graphdriver_btrfs,containers_image_openpgp" -o bin/$(BINARY_NAME) -v .

build-builder: ## Build sigma-builder and put the output binary in ./bin
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 GO111MODULE=on CC="$(CC)" CXX="$(CXX)" $(GOCMD) build $(GOFLAGS) -tags "timetzdata,exclude_graphdriver_devicemapper,exclude_graphdriver_btrfs,containers_image_openpgp" -o bin/$(BINARY_NAME)-builder -v ./cmd/builder

clean: ## Remove build related file
	rm -fr ./bin/sigma ./bin/sigma-builder ./bin/*.tar.gz ./bin/*.tar

vendor: ## Copy of all packages needed to support builds and tests in the vendor directory
	@$(GOCMD) mod tidy && $(GOCMD) mod vendor

## Lint:
lint: lint-go lint-dockerfile ## Run all available linters

.PHONY: lint-dockerfile
lint-dockerfile: ## Lint your Dockerfile
	@hadolint $(shell find build -name "*Dockerfile")

lint-go: ## Use golintci-lint on your project
	@golangci-lint run --timeout=10m --build-tags "timetzdata,exclude_graphdriver_devicemapper,exclude_graphdriver_btrfs,containers_image_openpgp"

## Docker:
docker-build: docker-build-builder-local dockerfile-local ## Use the dockerfile to build the sigma image
	@docker buildx build --build-arg USE_MIRROR=$(USE_MIRROR) --build-arg WITH_TRIVY_DB=$(WITH_TRIVY_DB) -f build/all.alpine.Dockerfile --platform $(DOCKER_PLATFORMS) --progress plain --output type=docker,name=$(DOCKER_REGISTRY)/$(BINARY_NAME):latest,push=false,oci-mediatypes=true,force-compression=true .

docker-build-builder: ## Use the dockerfile to build the sigma-builder image
	@docker buildx build --build-arg USE_MIRROR=$(USE_MIRROR) -f build/builder.Dockerfile --platform $(DOCKER_PLATFORMS) --progress plain --output type=docker,name=$(DOCKER_REGISTRY)/$(BINARY_NAME)-builder:latest,push=false,oci-mediatypes=true,force-compression=true .

docker-build-builder-local: ## Use the dockerfile to build the sigma-builder image and save to local tarball file
	@docker buildx build --build-arg USE_MIRROR=$(USE_MIRROR) -f build/builder.Dockerfile --platform linux/amd64,linux/arm64 --progress plain --output type=oci,name=$(DOCKER_REGISTRY)/$(BINARY_NAME)-builder:latest,push=false,oci-mediatypes=true,dest=./bin/builder.$(VERSION).tar .

dockerfile-local: ## Use skopeo to copy dockerfile to local tarball file
	@skopeo copy -a docker://docker/dockerfile:1.10.0 oci-archive:bin/dockerfile.1.10.0.tar

.PHONY: docker-build-web
docker-build-web: ## Build the web image
	@docker buildx build --build-arg USE_MIRROR=$(USE_MIRROR) -f build/web.Dockerfile --platform $(DOCKER_PLATFORMS) --progress plain --output type=docker,name=$(DOCKER_REGISTRY)/$(BINARY_NAME)-web:latest,push=false,oci-mediatypes=true,force-compression=true .

.PHONY: docker-build-trivy
docker-build-trivy: ## Build the trivy image
	@docker buildx build --build-arg USE_MIRROR=$(USE_MIRROR) -f build/trivy.Dockerfile --platform $(DOCKER_PLATFORMS) --progress plain --output type=docker,name=$(DOCKER_REGISTRY)/$(BINARY_NAME)-trivy:latest,push=false,oci-mediatypes=true,force-compression=true .

.PHONY: docker-build-local
docker-build-local: build ## Build the local sigma image
	@docker buildx build --build-arg USE_MIRROR=$(USE_MIRROR) --build-arg WITH_TRIVY_DB=$(WITH_TRIVY_DB) -f build/local.Dockerfile --platform $(DOCKER_PLATFORMS) --progress plain --output type=docker,name=$(DOCKER_REGISTRY)/$(BINARY_NAME):latest,push=false,oci-mediatypes=true,force-compression=true .

## Misc:
migration-create: ## Create a new migration file
	@migrate create -dir ./pkg/dal/migrations/mysql -seq -digits 4 -ext sql $(MIGRATION_NAME)

sql-format: ## Format all sql files
	@find ${PWD}/pkg -type f -iname "*.sql" -print | xargs pg_format -s 2 --inplace

changelog: ## Generate changelog
	@docker run -v "${PWD}":/workdir quay.io/git-chglog/git-chglog:latest --next-tag $(VERSION) -o CHANGELOG.md

gormgen: ## Generate gorm model from database
	@$(GOCMD) run ./pkg/dal/cmd/gen.go

swagen: ## Generate swagger from code comments
	# go install github.com/swaggo/swag/cmd/swag@latest
	@swag fmt
	@swag init --output pkg/handlers/apidocs

addlicense: ## Add license to all go files
	@find pkg -type f -name "*.go" | grep -v "pkg/handlers/apidocs/docs.go" | xargs addlicense -l apache -y 2024 -c "sigma"
	@find cmd -type f -name "*.go" | xargs addlicense -l apache -y 2024 -c "sigma"
	@addlicense -l apache -y 2024 -c "sigma" main.go
	@addlicense -l apache -y 2024 -c "sigma" web/web.go
	@find web/src -type f -name "*.tsx" | xargs addlicense -l apache -y 2024 -c "sigma"
	@find web/src -type f -name "*.ts" | xargs addlicense -l apache -y 2024 -c "sigma"
	@find web/src -type f -name "*.css" | xargs addlicense -l apache -y 2024 -c "sigma"

## Kube:
kube_install: ## Install sigma on k8s using helm
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

kube_uninstall: ## Uninstall sigma on k8s using helm
	@KUBECONFIG=$(KUBECONFIG)
	@helm uninstall $(APPNAME) -n$(NAMESPACE)

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-30s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

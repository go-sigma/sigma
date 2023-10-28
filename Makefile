GOCMD             = go
GOTEST            = $(GOCMD) test
GOVET             = $(GOCMD) vet
BINARY_NAME       = sigma
VERSION          ?= 0.0.0
SERVICE_PORT     ?= 3000
DOCKER_REGISTRY  ?= docker.io/tosone
DOCKER_PLATFORMS ?= linux/amd64,linux/arm64
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

GOLDFLAGS        += -X github.com/go-sigma/sigma/pkg/version.Version=$(shell git describe --tags --dirty --always)
GOLDFLAGS        += -X github.com/go-sigma/sigma/pkg/version.BuildDate=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOLDFLAGS        += -X github.com/go-sigma/sigma/pkg/version.GitHash=$(shell git rev-parse --short HEAD)
GOFLAGS           = -ldflags '-s -w $(GOLDFLAGS)'

.PHONY: all test build vendor

all: build build-builder

all-linux: build-linux build-builder-linux

## Build:
build: ## Build sigma and put the output binary in ./bin
	@GO111MODULE=on $(GOCMD) build $(GOFLAGS) -tags timetzdata -o bin/$(BINARY_NAME) -v .

build-builder: ## Build sigma-builder and put the output binary in ./bin
	@GO111MODULE=on $(GOCMD) build $(GOFLAGS) -tags timetzdata -o bin/$(BINARY_NAME)-builder -v ./cmd/builder

build-linux: ## Build sigma for linux and put the output binary in ./bin
	@GO111MODULE=on GOOS=linux $(GOCMD) build $(GOFLAGS) -tags timetzdata -o bin/$(BINARY_NAME) -v .

build-builder-linux: ## Build sigma-builder for release and put the output binary in ./bin
	@GO111MODULE=on GOOS=linux $(GOCMD) build $(GOFLAGS) -tags timetzdata -o bin/$(BINARY_NAME)-builder -v ./cmd/builder

clean: ## Remove build related file
	rm -fr ./bin
	rm -f ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.cov yamllint-checkstyle.xml

vendor: ## Copy of all packages needed to support builds and tests in the vendor directory
	@$(GOCMD) mod tidy && $(GOCMD) mod vendor

## Lint:
lint: lint-go lint-dockerfile lint-yaml ## Run all available linters

.PHONY: lint-dockerfile
lint-dockerfile: ## Lint your Dockerfile
# If dockerfile is present we lint it.
ifeq ($(shell test -e ./Dockerfile && echo -n yes),yes)
	$(eval CONFIG_OPTION = $(shell [ -e $(shell pwd)/.hadolint.yaml ] && echo "-v $(shell pwd)/.hadolint.yaml:/root/.config/hadolint.yaml" || echo "" ))
	$(eval OUTPUT_OPTIONS = $(shell [ "${EXPORT_RESULT}" == "true" ] && echo "--format checkstyle" || echo "" ))
	$(eval OUTPUT_FILE = $(shell [ "${EXPORT_RESULT}" == "true" ] && echo "| tee /dev/tty > checkstyle-report.xml" || echo "" ))
	docker run --rm -i $(CONFIG_OPTION) hadolint/hadolint hadolint $(OUTPUT_OPTIONS) - < ./Dockerfile $(OUTPUT_FILE)
endif

lint-go: ## Use golintci-lint on your project
	@golangci-lint run --deadline=10m

lint-yaml: ## Use yamllint on the yaml file of your projects
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/thomaspoignant/yamllint-checkstyle
	$(eval OUTPUT_OPTIONS = | tee /dev/tty | yamllint-checkstyle > yamllint-checkstyle.xml)
endif
	docker run --rm -it -v $(shell pwd):/data cytopia/yamllint -f parsable $(shell git ls-files '*.yml' '*.yaml') $(OUTPUT_OPTIONS)

## Docker:
docker-build: ## Use the dockerfile to build the container
	docker buildx build -f build/Dockerfile --platform $(DOCKER_PLATFORMS) --progress plain --output type=image,name=$(DOCKER_REGISTRY)/$(BINARY_NAME):latest,push=true .

docker-build-local: build-linux ## Build the container with the local binary
	docker buildx build -f build/Dockerfile.local --platform $(DOCKER_PLATFORMS) --progress plain --output type=image,name=$(DOCKER_REGISTRY)/$(BINARY_NAME):latest,push=true .

docker-build-builder: ## Build the dev container
	docker buildx build -f build/Dockerfile.builder --platform $(DOCKER_PLATFORMS) --progress plain --output type=image,name=$(DOCKER_REGISTRY)/$(BINARY_NAME)-builder:latest,push=true .

docker-build-builder-local: build-builder-linux # Build sigma builder image
	docker buildx build -f build/Dockerfile.builder.local --platform $(DOCKER_PLATFORMS) --progress plain --output type=image,name=$(DOCKER_REGISTRY)/$(BINARY_NAME)-builder:latest,push=true .

## Format:
format: sql-format

sql-format: ## format all sql files
	@find ${PWD}/pkg/dal/migrations -type f -iname "*.sql" -print | xargs pg_format -s 2 --inplace

## Misc:
migration-create: ## Create a new migration file
	@migrate create -dir ./migrations -seq -digits 4 -ext sql $(MIGRATION_NAME)

gormgen: ## Generate gorm model from database
	@$(GOCMD) run ./pkg/dal/cmd/gen.go

swagen: ## Generate swagger from code comments
	@swag fmt
	@swag init --output pkg/handlers/apidocs

addlicense: ## Add license to all go files
	@find pkg -type f -name "*.go" | xargs addlicense -l apache -y 2023 -c "sigma"
	@find cmd -type f -name "*.go" | xargs addlicense -l apache -y 2023 -c "sigma"
	@addlicense -l apache -y 2023 -c "sigma" main.go
	@addlicense -l apache -y 2023 -c "sigma" web/web.go
	@find web/src -type f -name "*.tsx" | xargs addlicense -l apache -y 2023 -c "sigma"
	@find web/src -type f -name "*.ts" | xargs addlicense -l apache -y 2023 -c "sigma"
	@find web/src -type f -name "*.css" | xargs addlicense -l apache -y 2023 -c "sigma"

## Kube:
kube_deploy: ## Deploy sigma on k8s using helm
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

kube_undeploy: ## Uninstall sigma on k8s using helm
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
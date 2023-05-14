GOCMD            = go
GOTEST           = $(GOCMD) test
GOVET            = $(GOCMD) vet
BINARY_NAME      = ximager
VERSION         ?= 0.0.0
SERVICE_PORT    ?= 3000
DOCKER_REGISTRY ?= #if set it should finished by /
EXPORT_RESULT   ?= false # for CI please set EXPORT_RESULT to true


MIGRATION_NAME  ?=

SHELL           := /bin/bash

GREEN           := $(shell tput -Txterm setaf 2)
YELLOW          := $(shell tput -Txterm setaf 3)
WHITE           := $(shell tput -Txterm setaf 7)
CYAN            := $(shell tput -Txterm setaf 6)
RESET           := $(shell tput -Txterm sgr0)

GOLDFLAGS       += -X github.com/ximager/ximager/cmd.version=$(shell git describe --tags --dirty)
GOLDFLAGS       += -X github.com/ximager/ximager/cmd.buildDate=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOLDFLAGS       += -X github.com/ximager/ximager/cmd.gitHash=$(shell git rev-parse --short HEAD)
GOFLAGS          = -ldflags '-extldflags "-static" -s -w $(GOLDFLAGS)'

.PHONY: all test build vendor

all: help

## Build:
build: ## Build your project and put the output binary in ./bin
	@$(GOCMD) mod download
	@CGO_ENABLED=0 GO111MODULE=on $(GOCMD) build $(GOFLAGS) -tags timetzdata -o bin/$(BINARY_NAME) -v .

build-release: ## Build your project for release and put the output binary in ./bin
	@$(GOCMD) mod download
	@CGO_ENABLED=0 GO111MODULE=on $(GOCMD) build $(GOFLAGS) -tags timetzdata -o bin/$(BINARY_NAME) -v .

build-linux: ## Build your project for linux and put the output binary in ./bin
	@CGO_ENABLED=0 GO111MODULE=on GOOS=linux $(GOCMD) build $(GOFLAGS) -tags timetzdata -o bin/$(BINARY_NAME) -v .

clean: ## Remove build related file
	rm -fr ./bin
	rm -f ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.cov yamllint-checkstyle.xml

vendor: ## Copy of all packages needed to support builds and tests in the vendor directory
	$(GOCMD) mod tidy && $(GOCMD) mod vendor

watch: ## Run the code with cosmtrek/air to have automatic reload on changes
	$(eval PACKAGE_NAME=$(shell head -n 1 go.mod | cut -d ' ' -f2))
	docker run -it --rm -w /go/src/$(PACKAGE_NAME) -v $(shell pwd):/go/src/$(PACKAGE_NAME) -p $(SERVICE_PORT):$(SERVICE_PORT) cosmtrek/air

## Test:
test: ## Run the tests of the project
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/jstemmer/go-junit-report
	$(eval OUTPUT_OPTIONS = | tee /dev/tty | go-junit-report -set-exit-code > junit-report.xml)
endif
	$(GOTEST) -v -race ./... $(OUTPUT_OPTIONS)

coverage: ## Run the tests of the project and export the coverage
	$(GOTEST) -cover -covermode=count -coverprofile=profile.cov ./...
	$(GOCMD) tool cover -func profile.cov
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/AlekSi/gocov-xml
	GO111MODULE=off go get -u github.com/axw/gocov/gocov
	gocov convert profile.cov | gocov-xml > coverage.xml
endif

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
	$(eval OUTPUT_OPTIONS = $(shell [ "${EXPORT_RESULT}" == "true" ] && echo "--out-format checkstyle ./... | tee /dev/tty > checkstyle-report.xml" || echo "" ))
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest-alpine golangci-lint run --deadline=10m $(OUTPUT_OPTIONS)

lint-yaml: ## Use yamllint on the yaml file of your projects
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/thomaspoignant/yamllint-checkstyle
	$(eval OUTPUT_OPTIONS = | tee /dev/tty | yamllint-checkstyle > yamllint-checkstyle.xml)
endif
	docker run --rm -it -v $(shell pwd):/data cytopia/yamllint -f parsable $(shell git ls-files '*.yml' '*.yaml') $(OUTPUT_OPTIONS)

## Docker:
docker-build: ## Use the dockerfile to build the container
	docker buildx build --platform=linux/amd64,linux/arm64 -f build/Dockerfile --rm --tag $(BINARY_NAME) .

docker-build-local: build-linux ## Build the container with the local binary
	docker build -f build/local.Dockerfile --rm --tag $(BINARY_NAME) .

docker-build-dev: ## Build the dev container
	docker build -f build/dev.Dockerfile --rm --tag $(BINARY_NAME) .

docker-release: ## Release the container with tag latest and version
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)
	# Push the docker images
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)

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
	@swag init --output pkg/handlers/apidocs

addlicense: ## Add license to all go files
	@find pkg -type f -name "*.go" | xargs addlicense -l apache -y 2023 -c "XImager" -v
	@find cmd -type f -name "*.go" | xargs addlicense -l apache -y 2023 -c "XImager"
	@addlicense -l apache -y 2023 -c "XImager" main.go
	@find web/src -type f -name "*.tsx" | xargs addlicense -l apache -y 2023 -c "XImager"
	@find web/src -type f -name "*.ts" | xargs addlicense -l apache -y 2023 -c "XImager"
	@find web/src -type f -name "*.css" | xargs addlicense -l apache -y 2023 -c "XImager"

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

include .env
export $(shell sed 's/=.*//' .env)

export COMPOSE_IGNORE_ORPHANS=True # ignore others container

PROJECT_NAME=sigma

## Misc:
.PHONY: install
install: ## Install docker
	curl -fsSL https://get.docker.com | sh
	sudo usermod -aG docker $(USER)

.PHONY: start
start: ## Start or restart all of the docker services
	if [ ! -d $(VOLUME_PREFIX)/traefik ]; then mkdir -p $(VOLUME_PREFIX)/traefik; fi
	if [ ! -f $(VOLUME_PREFIX)/traefik/acme.json ]; then touch $(VOLUME_PREFIX)/traefik/acme.json; fi
	chmod 600 $(VOLUME_PREFIX)/traefik/acme.json
	docker compose -p ${PROJECT_NAME} up -d

.PHONY: stop
stop: ## Stop all of the docker services
	docker compose -p ${PROJECT_NAME} down

## Help:
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

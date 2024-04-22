include .env
export

include deploy/build.mk

SHELL := /bin/bash -o pipefail

PROJECT_WORKSPACE := redify
DOCKER_COMPOSE := docker-compose -p $(PROJECT_WORKSPACE) -f deploy/develop/docker-compose.yml
DOCKER_BUILDKIT := 1
CONTAINER_IMAGE := demdxx/redify


.PHONY: generate-code
generate-code: ## Generate mocks for the project
	@echo "Generate mocks for the project"
	@go generate ./...


.PHONY: lint
lint: $(GOLANGLINTCI)
	golangci-lint run -v ./...


.PHONY: test
test: ## Run package test
	go test -tags ${APP_TAGS} -race ./...


.PHONY: tidy
tidy: ## Run mod tidy
	@echo "Run mod tidy"
	go mod tidy


.PHONY: godepup
godepup: ## Update current dependencies to the last version
	go get -u -v


.PHONY: fmt
fmt: ## Run formatting code
	@echo "Fix formatting"
	@gofmt -w ${GO_FMT_FLAGS} $$(go list -f "{{ .Dir }}" ./...); if [ "$${errors}" != "" ]; then echo "$${errors}"; fi


.PHONY: platform_list
platform_list:
	@echo $(DOCKER_PLATFORM_LIST)


.PHONY: run
run: ## Run current server
	@echo "Run current server"
	go run -tags ${APP_TAGS} cmd/main.go


.PHONY: build
build: ## Build application
	@echo "Build application"
	@rm -rf .build
	@$(call do_build,"cmd/main.go",redify)


.PHONY: build-docker-dev
build-docker-dev: build
	echo "Build develop docker image"
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build \
		-t ${CONTAINER_IMAGE}:latest -f deploy/develop/Dockerfile .


.PHONY: build-docker
build-docker: build ## Build production docker image
	@echo "Build docker image"
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker buildx build \
		--platform "${DOCKER_PLATFORM_LIST}" \
		-t ${CONTAINER_IMAGE}:${TAG_VALUE} \
		-t ${CONTAINER_IMAGE}:latest \
		-f deploy/production/Dockerfile .


.PHONY: run-srv
run-srv: build-docker-dev ## Run service by docker-compose
	@echo "Run service"
	$(DOCKER_COMPOSE) up redify


.PHONY: dbcli
dbcli: ## Open development database
	$(DOCKER_COMPOSE) exec pgdb psql -U $(DATABASE_USER) $(DATABASE_DB)


.PHONY: ch
ch: ## Run clickhouse client
	docker exec -it $(PROJECT_WORKSPACE)-clickhouse-1 clickhouse-client


.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.DEFAULT_GOAL := help

include .env
export

SHELL := /bin/bash -o pipefail

BUILD_GOOS ?= $(or ${DOCKER_DEFAULT_GOOS},linux)
BUILD_GOARCH ?= $(or ${DOCKER_DEFAULT_GOARCH},amd64)
# https://github.com/golang/go/wiki/MinimumRequirements#amd64
BUILD_GOAMD64 ?= $(or ${DOCKER_DEFAULT_GOAMD64},1)
BUILD_GOAMD64_LIST ?= $(or ${DOCKER_DEFAULT_GOAMD64_LIST},1)
BUILD_GOARM ?= $(or ${DOCKER_DEFAULT_GOARM},7)
BUILD_GOARM_LIST ?= 6 7
BUILD_CGO_ENABLED ?= 0
DOCKER_BUILDKIT ?= 1

LOCAL_TARGETPLATFORM=${BUILD_GOOS}/${BUILD_GOARCH}
ifeq (${BUILD_GOARCH},arm)
	LOCAL_TARGETPLATFORM=${BUILD_GOOS}/${BUILD_GOARCH}/v${BUILD_GOARM}
endif

COMMIT_NUMBER ?= $(or ${DEPLOY_COMMIT_NUMBER},)
ifeq (${COMMIT_NUMBER},)
	COMMIT_NUMBER = $(shell git log -1 --pretty=format:%h)
endif

TAG_VALUE ?= $(or ${DEPLOY_TAG_VALUE},)
ifeq (${TAG_VALUE},)
	TAG_VALUE = $(shell git describe --exact-match --tags `git log -n1 --pretty='%h'`)
endif
ifeq (${TAG_VALUE},)
	TAG_VALUE = commit-${COMMIT_NUMBER}
endif

OS_LIST   ?= $(or ${DEPLOY_OS_LIST},linux darwin)
ARCH_LIST ?= $(or ${DEPLOY_ARCH_LIST},amd64 arm64 arm)
APP_TAGS  ?= $(or ${APP_BUILD_TAGS},pgx)

export GO111MODULE := on
export PATH := $(GOBIN):$(PATH)
export GOSUMDB := off
export GOFLAGS=-mod=mod
# Go 1.13 defaults to TLS 1.3 and requires an opt-out.  Opting out for now until certs can be regenerated before 1.14
# https://golang.org/doc/go1.12#tls_1_3
export GODEBUG := tls13=0

define build_platform_list
	for os in $(OS_LIST); do \
		for arch in $(ARCH_LIST); do \
			if [ "$$os/$$arch" != "darwin/arm" ]; then \
				if [ "$$arch" = "arm" ]; then \
					for armv in $(BUILD_GOARM_LIST); do \
						i="$${os}/$${arch}/v$${armv}"; \
						echo -n "$${i},"; \
					done; \
				else \
					if [ "$$arch" = "amd64" ]; then \
						for amd64v in $(BUILD_GOAMD64_LIST); do \
							if [ "$$amd64v" == "1" ]; then \
								i="$${os}/$${arch}"; \
							else \
								i="$${os}/$${arch}/v$${amd64v}"; \
							fi; \
							echo -n "$${i},"; \
						done; \
					else \
						i="$${os}/$${arch}"; \
						echo -n "$${i},"; \
					fi; \
				fi; \
			fi; \
		done; \
	done;
endef

PROJECT_WORKSPACE := redify
DOCKER_COMPOSE := docker-compose -p $(PROJECT_WORKSPACE) -f deploy/develop/docker-compose.yml
DOCKER_BUILDKIT := 1
CONTAINER_IMAGE := demdxx/redify
DOCKER_PLATFORM_LIST := $(shell $(call build_platform_list))
DOCKER_PLATFORM_LIST := $(shell echo $(DOCKER_PLATFORM_LIST) | sed 's/.$$//')


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


define do_build
	@for os in $(OS_LIST); do \
		for arch in $(ARCH_LIST); do \
			if [ "$$os/$$arch" != "darwin/arm" ]; then \
				if [ "$$arch" = "arm" ]; then \
					for armv in $(BUILD_GOARM_LIST); do \
						echo "Build $$os/$$arch/v$$armv"; \
						GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} GOARM=$$armv \
							go build \
								-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
								-tags "${APP_TAGS}" -o .build/$$os/$$arch/v$$armv/$(2) $(1); \
					done; \
				else \
					if [ "$$arch" = "amd64" ]; then \
						for amd64v in $(BUILD_GOAMD64_LIST); do \
							if [ "$$amd64v" == "1" ]; then \
								echo "Build $$os/$$arch -> v1"; \
								GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} GOAMD64=v$$amd64v \
									go build \
										-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
										-tags "${APP_TAGS}" -o .build/$$os/$$arch/$(2) $(1); \
							else \
								echo "Build $$os/$$arch/v$$amd64v"; \
								GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} GOAMD64=v$$amd64v \
									go build \
										-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
										-tags "${APP_TAGS}" -o .build/$$os/$$arch/v$$amd64v/$(2) $(1); \
							fi; \
						done; \
					else \
						echo "Build $$os/$$arch"; \
						GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} \
							go build \
								-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
								-tags "${APP_TAGS}" -o .build/$$os/$$arch/$(2) $(1); \
					fi; \
				fi; \
			fi; \
		done; \
	done
endef


.PHONY: build
build: ## Build application
	@echo "Build application"
	@rm -rf .build
	@$(call do_build,"cmd/main.go",redify)


.PHONY: build-docker-dev
build-docker-dev: build
	echo "Build develop docker image"
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build \
		--build-arg TARGETPLATFORM=${LOCAL_TARGETPLATFORM} \
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

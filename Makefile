include .env
export

SHELL := /bin/bash -o pipefail
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

BUILD_GOOS ?= $(or ${DOCKER_DEFAULT_GOOS},linux)
BUILD_GOARCH ?= $(or ${DOCKER_DEFAULT_GOARCH},amd64)
BUILD_GOARM ?= 7
BUILD_CGO_ENABLED ?= 0

COMMIT_NUMBER ?= $(shell git log -1 --pretty=format:%h)
TAG_VALUE ?= `git describe --exact-match --tags $(git log -n1 --pretty='%h')`

PROJECT_WORKSPACE := redify

DOCKER_COMPOSE := docker-compose -p $(PROJECT_WORKSPACE) -f docker/docker-compose.yml
DOCKER_BUILDKIT := 1

APP_TAGS := "$(or ${APP_BUILD_TAGS},pgx)"

export GO111MODULE := on
export PATH := $(GOBIN):$(PATH)
export GOSUMDB := off
export GOFLAGS=-mod=mod
# Go 1.13 defaults to TLS 1.3 and requires an opt-out.  Opting out for now until certs can be regenerated before 1.14
# https://golang.org/doc/go1.12#tls_1_3
export GODEBUG := tls13=0


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

.PHONY: run
run: ## Run current server
	@echo "Run current server"
	go run -tags ${APP_TAGS} cmd/main.go

.PHONY: build
build: ## Build application
	@echo "Build application"
	@rm -rf .build/redify
	GOOS=${BUILD_GOOS} GOARCH=${BUILD_GOARCH} CGO_ENABLED=${BUILD_CGO_ENABLED} GOARM=${BUILD_GOARM} \
		go build -ldflags "-X main.buildDate=`date -u +%Y%m%d.%H%M%S` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE}" \
			-tags ${APP_TAGS} -o ".build/redify" cmd/main.go
	file .build/redify

CONTAINER_IMAGE := demdxx/redify:latest

.PHONY: build-docker-dev
build-docker-dev: build
	echo "Build develop docker image"
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build -t ${CONTAINER_IMAGE} -f docker/Dockerfile .

.PHONY: run-srv
run-srv: build-docker-dev ## Run service by docker-compose
	@echo "Run service"
	$(DOCKER_COMPOSE) up redify

.PHONY: dbcli
dbcli: ## Open development database
	$(DOCKER_COMPOSE) exec pgdb psql -U $(DATABASE_USER) $(DATABASE_DB)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

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

TMP_BASE := .tmp
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
TMP_BIN = $(TMP)/bin
TMP_ETC := $(TMP)/etc
TMP_LIB := $(TMP)/lib
TMP_VERSIONS := $(TMP)/versions

DOCKER_COMPOSE := docker-compose -p $(PROJECT_WORKSPACE) -f docker/docker-compose.yml
DOCKER_BUILDKIT := 1

APP_TAGS := "$(or ${APP_BUILD_TAGS},pgx)"

unexport GOPATH
export GOPATH=$(abspath $(TMP))
export GO111MODULE := on
export GOBIN := $(abspath $(TMP_BIN))
export PATH := $(GOBIN):$(PATH)
export GOSUMDB := off
export GOFLAGS=-mod=mod
# Go 1.13 defaults to TLS 1.3 and requires an opt-out.  Opting out for now until certs can be regenerated before 1.14
# https://golang.org/doc/go1.12#tls_1_3
export GODEBUG := tls13=0

GOLANGLINTCI_VERSION := latest
GOLANGLINTCI := $(TMP_VERSIONS)/golangci-lint/$(GOLANGLINTCI_VERSION)
$(GOLANGLINTCI):
	$(eval GOLANGLINTCI_TMP := $(shell mktemp -d))
	cd $(GOLANGLINTCI_TMP); go get github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGLINTCI_VERSION)
	@rm -rf $(GOLANGLINTCI_TMP)
	@rm -rf $(dir $(GOLANGLINTCI))
	@mkdir -p $(dir $(GOLANGLINTCI))
	@touch $(GOLANGLINTCI)

ERRCHECK_VERSION := v1.2.0
ERRCHECK := $(TMP_VERSIONS)/errcheck/$(ERRCHECK_VERSION)
$(ERRCHECK):
	$(eval ERRCHECK_TMP := $(shell mktemp -d))
	cd $(ERRCHECK_TMP); go get github.com/kisielk/errcheck@$(ERRCHECK_VERSION)
	@rm -rf $(ERRCHECK_TMP)
	@rm -rf $(dir $(ERRCHECK))
	@mkdir -p $(dir $(ERRCHECK))
	@touch $(ERRCHECK)

# UPDATE_LICENSE_VERSION := ce2550dad7144b81ae2f67dc5e55597643f6902b
# UPDATE_LICENSE := $(TMP_VERSIONS)/update-license/$(UPDATE_LICENSE_VERSION)
# $(UPDATE_LICENSE):
# 	$(eval UPDATE_LICENSE_TMP := $(shell mktemp -d))
# 	cd $(UPDATE_LICENSE_TMP); go get go.uber.org/tools/update-license@$(UPDATE_LICENSE_VERSION)
# 	@rm -rf $(UPDATE_LICENSE_TMP)
# 	@rm -rf $(dir $(UPDATE_LICENSE))
# 	@mkdir -p $(dir $(UPDATE_LICENSE))
# 	@touch $(UPDATE_LICENSE)

CERTSTRAP_VERSION := v1.1.1
CERTSTRAP := $(TMP_VERSIONS)/certstrap/$(CERTSTRAP_VERSION)
$(CERTSTRAP):
	$(eval CERTSTRAP_TMP := $(shell mktemp -d))
	cd $(CERTSTRAP_TMP); go get github.com/square/certstrap@$(CERTSTRAP_VERSION)
	@rm -rf $(CERTSTRAP_TMP)
	@rm -rf $(dir $(CERTSTRAP))
	@mkdir -p $(dir $(CERTSTRAP))
	@touch $(CERTSTRAP)

GOMOCK_VERSION := v1.3.1
GOMOCK := $(TMP_VERSIONS)/mockgen/$(GOMOCK_VERSION)
$(GOMOCK):
	$(eval GOMOCK_TMP := $(shell mktemp -d))
	cd $(GOMOCK_TMP); go get github.com/golang/mock/mockgen@$(GOMOCK_VERSION)
	@rm -rf $(GOMOCK_TMP)
	@rm -rf $(dir $(GOMOCK))
	@mkdir -p $(dir $(GOMOCK))
	@touch $(GOMOCK)

.PHONY: deps
deps: $(GOLANGLINTCI) $(ERRCHECK) $(CERTSTRAP) $(GOMOCK)

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

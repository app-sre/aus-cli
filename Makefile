# Ensure go modules are enabled:
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

# Docker setup
IMAGE_NAME := quay.io/app-sre/ocm-aus-cli
IMAGE_TAG := $(shell git rev-parse --short=7 HEAD)

ifneq (,$(wildcard $(CURDIR)/.docker))
	DOCKER_CONF := $(CURDIR)/.docker
else
	DOCKER_CONF := $(HOME)/.docker
endif

# Disable CGO so that we always generate static binaries:
export CGO_ENABLED=0

# Allow overriding: `make lint CONTAINER_ENGINE=docker`.
CONTAINER_ENGINE ?= $(shell which podman >/dev/null 2>&1 && echo podman || echo docker)


.PHONY: all
all: build

.PHONY: build
build:
	go build -ldflags "-X github.com/app-sre/aus-cli/cmd/ocm-aus/version.Version=`git describe --tags --abbrev=0` -X github.com/app-sre/aus-cli/cmd/ocm-aus/version.Commit=`git rev-parse HEAD`" ./cmd/ocm-aus

.PHONY: release
release:
	@$(CONTAINER_ENGINE) run --rm -v "$(PWD):/app" -u $(id -u ${USER}):$(id -g ${USER}) -e GITHUB_TOKEN=$(GITHUB_TOKEN) --workdir=/app ghcr.io/goreleaser/goreleaser:v1.18.2 release

.PHONY: test
test: build
	CGO_ENABLED=0 GOOS=$(GOOS) go test ./...

.PHONY: fmt
fmt:
	gofmt -s -l -w cmd pkg

.PHONY: image-image
build-image:
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) build -t quay.io/app-sre/ocm-aus-cli:latest . --progress=plain

.PHONY: image-push
push-image: build-image
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push $(IMAGE_NAME):latest
	@$(CONTAINER_ENGINE) tag $(IMAGE_NAME):latest $(IMAGE_NAME):$(IMAGE_TAG)
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push $(IMAGE_NAME):$(IMAGE_TAG)

.PHONY: lint
lint:
	@$(CONTAINER_ENGINE) run --rm -w app -v "$(PWD):/app" --workdir=/app \
		quay.io/app-sre/golangci-lint:v$(shell cat .golangciversion) \
		golangci-lint run --timeout 15m

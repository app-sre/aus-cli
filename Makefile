# Ensure go modules are enabled:
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

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
	@$(CONTAINER_ENGINE) run --rm -v "$(PWD):/app:z" -u $(id -u ${USER}):$(id -g ${USER}) -e GITHUB_TOKEN=$(GITHUB_TOKEN) --workdir=/app ghcr.io/goreleaser/goreleaser:v1.18.2 release

.PHONY: test
test: build
	CGO_ENABLED=0 GOOS=$(GOOS) go test ./...

.PHONY: fmt
fmt:
	gofmt -s -l -w cmd pkg

.PHONY: lint
lint:
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) run --rm -w app -v "$(PWD):/app:z" --workdir=/app \
		quay.io/app-sre/golangci-lint:v2.3.0 \
		golangci-lint run

# Ensure go modules are enabled:
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

# Disable CGO so that we always generate static binaries:
export CGO_ENABLED=0

# Allow overriding: `make lint container_runner=docker`.
container_runner ?= $(shell which podman >/dev/null 2>&1 && echo podman || echo docker)


.PHONY: all
all: build

.PHONY: build
build:
	go build ./cmd/ocm-aus

.PHONY: release
release:
	$(container_runner) run --rm -v "$(PWD):/app" -u $(id -u ${USER}):$(id -g ${USER}) -e GITHUB_TOKEN=$(GITHUB_TOKEN) --workdir=/app ghcr.io/goreleaser/goreleaser:v1.18.2 release

.PHONY: test
test: build
	CGO_ENABLED=0 GOOS=$(GOOS) go test ./...

.PHONY: fmt
fmt:
	gofmt -s -l -w cmd pkg

.PHONY: lint
lint:
	$(container_runner) run --rm -w app -v "$(PWD):/app" --workdir=/app \
		quay.io/app-sre/golangci-lint:v$(shell cat .golangciversion) \
		golangci-lint run --timeout 15m

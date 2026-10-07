FROM registry.access.redhat.com/ubi9/go-toolset:9.8-1791275853@sha256:890b54e8d329f33f094ab17881bd37e20b80a54cc39141637310911d8dd9c575 AS base
COPY LICENSE /licenses/LICENSE
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .

FROM base AS builder
RUN make build

FROM base AS test
COPY --from=quay.io/app-sre/golangci-lint:v2.3.0 /usr/bin/golangci-lint /bin/golangci-lint
RUN golangci-lint run
RUN make test

FROM quay.io/redhat-services-prod/openshift/ocm-container:8ad42b3@sha256:dd9e2bb44c69c123b53c5ed61377bc9b4fd94385a331de79dd96aa94be839d57
COPY --from=builder /build/ocm-aus /usr/local/bin

FROM registry.access.redhat.com/ubi9/go-toolset:1.26.10-1792671270@sha256:edb9452603fa48e9882063f001ecd7d322add20f1564a4f05f38f6683be5ea94 AS base
COPY LICENSE /licenses/LICENSE
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .

FROM base AS builder
RUN make build

FROM base AS test
COPY --from=quay.io/app-sre/golangci-lint:v2.15.0@sha256:f4ea5f12ca354cc8682c1c3565c358915ff9e49f7ff7aea4e253f23f1e8422dc /usr/bin/golangci-lint /bin/golangci-lint
RUN golangci-lint run
RUN make test

FROM quay.io/redhat-services-prod/openshift/ocm-container:8ad42b3@sha256:dd9e2bb44c69c123b53c5ed61377bc9b4fd94385a331de79dd96aa94be839d57
COPY --from=builder /build/ocm-aus /usr/local/bin

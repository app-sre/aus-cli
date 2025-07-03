FROM registry.access.redhat.com/ubi9/go-toolset:1.23.9-1751538372@sha256:381fb72f087a07432520fa93364f66b5981557f1dd708f3c4692d6d0a76299b3 AS builder
COPY LICENSE /licenses/LICENSE
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM builder as test
RUN make test

FROM quay.io/redhat-services-prod/openshift/ocm-container:8ad42b3
COPY --from=builder /build/ocm-aus /usr/local/bin

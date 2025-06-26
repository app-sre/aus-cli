FROM registry.access.redhat.com/ubi9/go-toolset:1.23.9-1749636489@sha256:2a88121395084eaa575e5758b903fffb43dbf9d9586b2878e51678f63235b587 AS builder

WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM quay.io/app-sre/ocm-container:6d322fb

COPY --from=builder /build/ocm-aus /usr/local/bin

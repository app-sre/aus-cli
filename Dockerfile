FROM registry.access.redhat.com/ubi9/go-toolset:9.6-1750969886@sha256:3bbd87d77ea93742bd71a5275a31ec4a7693454ab80492c6a7d28ce6eef35378 AS builder
COPY LICENSE /licenses/LICENSE
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM builder as test
RUN make test

FROM quay.io/app-sre/ocm-container:6d322fb
COPY --from=builder /build/ocm-aus /usr/local/bin

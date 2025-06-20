FROM registry.access.redhat.com/ubi9/go-toolset:1.18.10-4.1683015641 AS builder

WORKDIR /build
COPY . .
RUN make build

FROM registry.access.redhat.com/ubi9-minimal:9.6-1749489516@sha256:9bed53318702feb9b15c79d56d4fc2a6857fdffa335eee7be53421989c7658d1

USER 1001

COPY --from=builder --chown=1001:0 /build/ocm-aus /usr/local/bin

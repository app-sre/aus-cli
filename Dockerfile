FROM registry.access.redhat.com/ubi9/go-toolset:1.18.10-4.1683015641 AS builder

WORKDIR /build
COPY . .
RUN make build

FROM registry.access.redhat.com/ubi9-minimal:9.5-1733767867@sha256:dee813b83663d420eb108983a1c94c614ff5d3fcb5159a7bd0324f0edbe7fca1

COPY --from=builder /build/ocm-aus /usr/local/bin

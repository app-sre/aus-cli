FROM registry.access.redhat.com/ubi9/go-toolset:1.23.9-1749636489@sha256:2a88121395084eaa575e5758b903fffb43dbf9d9586b2878e51678f63235b587 AS builder
COPY LICENSE /licenses/LICENSE
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM builder as test
RUN make test

FROM quay.io/redhat-services-prod/openshift/ocm-container:8ad42b3@sha256:dd9e2bb44c69c123b53c5ed61377bc9b4fd94385a331de79dd96aa94be839d57
COPY --from=builder /build/ocm-aus /usr/local/bin

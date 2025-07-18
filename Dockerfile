FROM registry.access.redhat.com/ubi9/go-toolset:9.6-1752083840@sha256:6fd64cd7f38a9b87440f963b6c04953d04de65c35b9672dbd7f1805b0ae20d09 AS builder
COPY LICENSE /licenses/LICENSE
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM builder as test
RUN make test

FROM quay.io/redhat-services-prod/openshift/ocm-container:8ad42b3@sha256:dd9e2bb44c69c123b53c5ed61377bc9b4fd94385a331de79dd96aa94be839d57
COPY --from=builder /build/ocm-aus /usr/local/bin

FROM quay.io/app-sre/golang:1.18.10 as builder

WORKDIR /build
COPY . .
RUN make build

FROM quay.io/app-sre/ocm-container:6d322fb

COPY --from=builder /build/ocm-aus /usr/local/bin

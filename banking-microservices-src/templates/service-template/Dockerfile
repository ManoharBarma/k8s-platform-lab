# syntax=docker/dockerfile:1
#
# Multi-stage build for a banking-platform microservice.
# IMPORTANT: build with the repository ROOT as the build context so the shared
# module (pkg/) is available, e.g.:
#   docker build -f templates/service-template/Dockerfile -t service-template .
#
# The distroless "nonroot" final image runs as a non-root user with no shell,
# minimizing the attack surface.

# ---- build stage ----
FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /src

# Copy the shared module and this service, preserving the relative layout so the
# `replace ../../pkg` directive in go.mod resolves.
COPY pkg/ ./pkg/
COPY templates/service-template/ ./templates/service-template/

WORKDIR /src/templates/service-template
RUN go mod download
RUN go build -trimpath -ldflags="-s -w" -o /out/app ./cmd

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=builder /out/app /app

USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/app"]

# path: docker/base/Dockerfile.go
# Base image for Go services (compilation stage).
# Centralizes the toolchain version and cache mounts.
FROM golang:1.26-alpine AS go-builder
RUN apk add --no-cache git ca-certificates make
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

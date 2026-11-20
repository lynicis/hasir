FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /usr/src/app

COPY go.work go.work.sum ./
COPY apps/api/go.mod apps/api/go.sum ./apps/api/
COPY packages/proto/go.mod packages/proto/go.sum ./packages/proto/

# Download dependencies for all modules in workspace
RUN go mod download

COPY apps/api ./apps/api
COPY packages/proto ./packages/proto

WORKDIR /usr/src/app/apps/api
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o api .

FROM alpine:3.24

WORKDIR /usr/src/app

COPY --from=build /usr/src/app/apps/api/api .
COPY apps/api/migrations ./migrations

RUN apk upgrade --no-cache \
  && apk add --no-cache git ca-certificates
RUN addgroup -S hasir && adduser -S -G hasir hasir
RUN chown -R hasir:hasir /usr/src/app

USER hasir

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
  CMD wget -q --spider "http://127.0.0.1:${PORT:-8080}/healthz" || exit 1

ENTRYPOINT ["/usr/src/app/api"]

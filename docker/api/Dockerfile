# path: docker/api/Dockerfile
# Build runs in the context of the repository root, so `apps/api`
# and `proto/gen` are both available to the Go compiler.
# 
# See docker/base/Dockerfile.go for the base builder image.

FROM hasir-go-base:latest AS builder
WORKDIR /app

# Copy go.mod and go.sum from api and root
COPY apps/api/go.mod apps/api/go.sum ./apps/api/
# Copy the generated protobuf output so the replace directive resolves
COPY proto/gen/go ./proto/gen/go

WORKDIR /app/apps/api
RUN go mod download

# Copy source
COPY apps/api/ ./
# Build the binary
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o /api main.go

# ---
FROM alpine:latest
WORKDIR /usr/src/app

RUN apk add --no-cache git ca-certificates protoc npm go \
    && npm i -g @bufbuild/protoc-gen-es @connectrpc/protoc-gen-connect-es \
    && addgroup -S hasir && adduser -S -G hasir hasir

COPY --from=builder /api ./api
RUN chown -R hasir:hasir /usr/src/app

USER hasir
ENTRYPOINT ["/usr/src/app/api"]

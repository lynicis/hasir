FROM golang:1.26-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /usr/src/app

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o api .

FROM alpine:latest

WORKDIR /usr/src/app

COPY --from=build /usr/src/app/api .

RUN apk add --no-cache git ca-certificates
RUN addgroup -S hasir && adduser -S -G hasir hasir
RUN chown -R hasir:hasir /usr/src/app

USER hasir

ENTRYPOINT ["/usr/src/app/api"]

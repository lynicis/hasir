FROM alpine:latest

WORKDIR /usr/src/app

COPY api ./

RUN apk add --no-cache git ca-certificates protoc npm go

RUN addgroup -S hasir && adduser -S -G hasir hasir
RUN chown -R hasir:hasir /usr/src/app

USER hasir

ENTRYPOINT ["/usr/src/app/api"]

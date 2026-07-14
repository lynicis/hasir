FROM alpine:latest

WORKDIR /usr/src/app

COPY api ./

RUN apk add --no-cache git ca-certificates protoc npm go
RUN npm i -g @bufbuild/protoc-gen-es
RUN npm i -g @connectrpc/protoc-gen-connect-es

RUN addgroup -S hasir && adduser -S -G hasir hasir
RUN chown -R hasir:hasir /usr/src/app

USER hasir

ENTRYPOINT ["/usr/src/app/api"]

FROM golang:1.25.3-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux go build -o api

FROM golang:1.25.3-alpine

WORKDIR /usr/src/app

COPY --from=build /build/api .

RUN adduser -D hasir
RUN chown -R hasir:hasir /usr/src/app

USER hasir

ENTRYPOINT ["/usr/src/app/api"]
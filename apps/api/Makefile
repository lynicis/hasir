.PHONY: build dev generate-mocks lint lint-fix run-postgres sec test vuln

BASE_PKG := hasir-api

build:
	go build ./...

dev:
	MODE=development go run main.go

test:
	go test -count=1 -coverprofile=coverage.out -covermode=atomic ./...

generate-mocks:
	go tool mockgen -package=user -destination=internal/user/service_mock.go -source=internal/user/service.go
	go tool mockgen -package=user -destination=internal/user/repository_mock.go -source=internal/user/repository.go
	go tool mockgen -package=registry -destination=internal/registry/service_mock.go -source=internal/registry/service.go
	go tool mockgen -package=registry -destination=internal/registry/repository_mock.go -source=internal/registry/repository.go
	go tool mockgen -package=registry -destination=internal/registry/queue_mock.go -source=internal/registry/queue.go
	go tool mockgen -package=organization -destination=internal/organization/service_mock.go -source=internal/organization/service.go
	go tool mockgen -package=organization -destination=internal/organization/repository_mock.go -source=internal/organization/repository.go
	go tool mockgen -package=organization -destination=internal/organization/queue_mock.go -source=internal/organization/queue.go
	go tool mockgen -package=email -destination=pkg/email/email_mock.go -source=pkg/email/email.go
	go tool mockgen -package=authorization -destination=pkg/authorization/authorization_mock.go -source=pkg/authorization/authorization.go
	go tool mockgen -package=idempotency -destination=pkg/idempotency/interceptor_mock.go -source=pkg/idempotency/interceptor.go IdempotencyRepository

lint:
	go tool golangci-lint run ./...

lint-fix:
	go tool golangci-lint run --fix ./...

run-postgres:
	docker run -p 5432:5432 --name postgres -d -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test postgres:alpine

sec:
	go tool gosec -exclude-dir=$(BASE_PKG)/sdk -exclude-dir=$(BASE_PKG)/repos ./...

vuln:
	go tool govulncheck ./...

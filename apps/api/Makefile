.PHONY: build dev generate-mocks lint lint-fix run-postgres sec test vuln

BASE_PKG := hasir-api
mockgen := "go tool mockgen"
golangci-lint := "go tool golangci-lint"
gosec := "go tool gosec"
govulncheck := "go tool govulncheck"

build:
	go build ./...

dev:
	MODE=development go run main.go

test:
	go test -count=1 -coverprofile=coverage.out -covermode=atomic ./...

generate-mocks:
	mockgen -package=user -destination=internal/user/service_mock.go -source=internal/user/service.go
	mockgen -package=user -destination=internal/user/repository_mock.go -source=internal/user/repository.go
	mockgen -package=registry -destination=internal/registry/service_mock.go -source=internal/registry/service.go
	mockgen -package=registry -destination=internal/registry/repository_mock.go -source=internal/registry/repository.go
	mockgen -package=registry -destination=internal/registry/queue_mock.go -source=internal/registry/queue.go
	mockgen -package=organization -destination=internal/organization/service_mock.go -source=internal/organization/service.go
	mockgen -package=organization -destination=internal/organization/repository_mock.go -source=internal/organization/repository.go
	mockgen -package=organization -destination=internal/organization/queue_mock.go -source=internal/organization/queue.go
	mockgen -package=email -destination=pkg/email/email_mock.go -source=pkg/email/email.go
	mockgen -package=authorization -destination=pkg/authorization/authorization_mock.go -source=pkg/authorization/authorization.go
	mockgen -package=idempotency -destination=pkg/idempotency/interceptor_mock.go -source=pkg/idempotency/interceptor.go IdempotencyRepository
	mockgen -package=authorization -destination=pkg/authorization/adapter_mock.go -source=pkg/authorization/adapter.go OrganizationRepository

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

run-postgres:
	docker run -p 5432:5432 --name postgres -d -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test postgres:alpine

sec:
	gosec -exclude-dir=$(BASE_PKG)/sdk -exclude-dir=$(BASE_PKG)/repos ./...

vuln:
	govulncheck ./...

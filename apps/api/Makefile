.PHONY: build dev generate-mocks lint lint-fix run-postgres sec test vuln

MOCKGEN := go tool mockgen
GOLANGCI_LINT := go tool golangci-lint
GOSEC := go tool gosec
GOVULNCHECK := go tool govulncheck

build:
	go build ./...

dev:
	MODE=development go run main.go

test:
	go test -count=1 -coverprofile=coverage.out -covermode=atomic ./...

generate-mocks:
	$(MOCKGEN) -package=user -destination=internal/user/service_mock.go -source=internal/user/service.go
	$(MOCKGEN) -package=user -destination=internal/user/repository_mock.go -source=internal/user/repository.go
	$(MOCKGEN) -package=registry -destination=internal/registry/service_mock.go -source=internal/registry/service.go
	$(MOCKGEN) -package=registry -destination=internal/registry/repository_mock.go -source=internal/registry/repository.go
	$(MOCKGEN) -package=registry -destination=internal/registry/queue_mock.go -source=internal/registry/queue.go
	$(MOCKGEN) -package=organization -destination=internal/organization/service_mock.go -source=internal/organization/service.go
	$(MOCKGEN) -package=organization -destination=internal/organization/repository_mock.go -source=internal/organization/repository.go
	$(MOCKGEN) -package=organization -destination=internal/organization/queue_mock.go -source=internal/organization/queue.go
	$(MOCKGEN) -package=email -destination=pkg/email/email_mock.go -source=pkg/email/email.go
	$(MOCKGEN) -package=authorization -destination=pkg/authorization/authorization_mock.go -source=pkg/authorization/authorization.go
	$(MOCKGEN) -package=idempotency -destination=pkg/idempotency/interceptor_mock.go -source=pkg/idempotency/interceptor.go IdempotencyRepository
	$(MOCKGEN) -package=authorization -destination=pkg/authorization/adapter_mock.go -source=pkg/authorization/adapter.go OrganizationRepository

lint:
	$(GOLANGCI_LINT) run ./...

lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

run-postgres:
	docker run -p 5432:5432 --name postgres -d -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test postgres:alpine

sec:
	$(GOSEC) . ./internal/... ./pkg/...

vuln:
	$(GOVULNCHECK) ./...

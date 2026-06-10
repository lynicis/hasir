.PHONY: dev generate-mocks lint lint-fix run-postgres sec

BASE_PKG := hasir-api

dev:
	MODE=development go run main.go

generate-mocks:
	# User
	mockgen -package=user -destination=internal/user/service_mock.go $(BASE_PKG)/internal/user Service
	mockgen -package=user -destination=internal/user/repository_mock.go $(BASE_PKG)/internal/user Repository
	# Registry
	mockgen -package=registry -destination=internal/registry/service_mock.go $(BASE_PKG)/internal/registry Service
	mockgen -package=registry -destination=internal/registry/repository_mock.go $(BASE_PKG)/internal/registry Repository
	mockgen -package=registry -destination=internal/registry/queue_mock.go $(BASE_PKG)/internal/registry SdkGenerationQueue
	# Organization
	mockgen -package=organization -destination=internal/organization/service_mock.go $(BASE_PKG)/internal/organization Service
	mockgen -package=organization -destination=internal/organization/repository_mock.go $(BASE_PKG)/internal/organization Repository
	mockgen -package=organization -destination=internal/organization/queue_mock.go $(BASE_PKG)/internal/organization Queue
	# Cross-cutting
	mockgen -package=email -destination=pkg/email/email_mock.go $(BASE_PKG)/pkg/email Service
	mockgen -package=authorization -destination=pkg/authorization/authorization_mock.go $(BASE_PKG)/pkg/authorization MemberRoleChecker

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

run-postgres:
	docker run -p 5432:5432 --name postgres -d -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test postgres:alpine

sec:
	gosec -exclude-dir=$(BASE_PKG)/sdk -exclude-dir=$(BASE_PKG)/repos ./...

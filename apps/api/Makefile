.PHONY: build dev generate-mocks lint lint-fix run-postgres sec test vuln

BASE_PKG := hasir-api

build:
	go build ./...

dev:
	MODE=development go run main.go

test:
	go test -count=1 -coverprofile=coverage.out -covermode=atomic ./...

generate-mocks:
	# User
	go run go.uber.org/mock/mockgen -package=user -destination=internal/user/service_mock.go $(BASE_PKG)/internal/user Service
	go run go.uber.org/mock/mockgen -package=user -destination=internal/user/repository_mock.go $(BASE_PKG)/internal/user Repository
	# Registry
	go run go.uber.org/mock/mockgen -package=registry -destination=internal/registry/service_mock.go $(BASE_PKG)/internal/registry Service
	go run go.uber.org/mock/mockgen -package=registry -destination=internal/registry/repository_mock.go $(BASE_PKG)/internal/registry Repository
	go run go.uber.org/mock/mockgen -package=registry -destination=internal/registry/queue_mock.go $(BASE_PKG)/internal/registry SdkGenerationQueue
	# Organization
	go run go.uber.org/mock/mockgen -package=organization -destination=internal/organization/service_mock.go $(BASE_PKG)/internal/organization Service
	go run go.uber.org/mock/mockgen -package=organization -destination=internal/organization/repository_mock.go $(BASE_PKG)/internal/organization Repository
	go run go.uber.org/mock/mockgen -package=organization -destination=internal/organization/queue_mock.go $(BASE_PKG)/internal/organization Queue
	# Cross-cutting
	go run go.uber.org/mock/mockgen -package=email -destination=pkg/email/email_mock.go $(BASE_PKG)/pkg/email Service
	go run go.uber.org/mock/mockgen -package=authorization -destination=pkg/authorization/authorization_mock.go $(BASE_PKG)/pkg/authorization MemberRoleChecker

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...

lint-fix:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix ./...

run-postgres:
	docker run -p 5432:5432 --name postgres -d -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test postgres:alpine

sec:
	go run github.com/securego/gosec/v2/cmd/gosec -exclude-dir=$(BASE_PKG)/sdk -exclude-dir=$(BASE_PKG)/repos ./...

vuln:
	go run golang.org/x/vuln/cmd/govulncheck ./...

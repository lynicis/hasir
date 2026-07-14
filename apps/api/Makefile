# path: apps/api/Makefile
# Turborepo delegates to these targets via package.json.

BASE_PKG := hasir-api

.PHONY: dev build test lint lint-fix generate-mocks clean sec vuln run-postgres

dev:
	MODE=development go run main.go

build:
	GOOS=linux GOARCH=amd64 go build -o api ./main.go

test:
	go test -v -coverprofile=coverage.txt ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

generate-mocks:
	mockgen -package=user -destination=internal/user/service_mock.go $(BASE_PKG)/internal/user Service
	mockgen -package=user -destination=internal/user/repository_mock.go $(BASE_PKG)/internal/user Repository
	mockgen -package=registry -destination=internal/registry/service_mock.go $(BASE_PKG)/internal/registry Service
	mockgen -package=registry -destination=internal/registry/repository_mock.go $(BASE_PKG)/internal/registry Repository
	mockgen -package=registry -destination=internal/registry/queue_mock.go $(BASE_PKG)/internal/registry SdkGenerationQueue
	mockgen -package=organization -destination=internal/organization/service_mock.go $(BASE_PKG)/internal/organization Service
	mockgen -package=organization -destination=internal/organization/repository_mock.go $(BASE_PKG)/internal/organization Repository
	mockgen -package=organization -destination=internal/organization/queue_mock.go $(BASE_PKG)/internal/organization Queue
	mockgen -package=email -destination=pkg/email/email_mock.go $(BASE_PKG)/pkg/email Service
	mockgen -package=authorization -destination=pkg/authorization/authorization_mock.go $(BASE_PKG)/pkg/authorization MemberRoleChecker

clean:
	rm -f api coverage.txt

sec:
	gosec -exclude-dir=$(BASE_PKG)/sdk -exclude-dir=$(BASE_PKG)/repos ./...

vuln:
	govulncheck ./...

run-postgres:
	docker run -p 5432:5432 --name postgres -d -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test postgres:16-alpine

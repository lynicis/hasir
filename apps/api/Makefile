generate-mocks:
	mockgen -package=user -destination=internal/user/service_mock.go apps/api/internal/user Service
	mockgen -package=user -destination=internal/user/repository_mock.go apps/api/internal/user Repository
	mockgen -package=registry -destination=internal/registry/service_mock.go apps/api/internal/registry Service
	mockgen -package=registry -destination=internal/registry/repository_mock.go apps/api/internal/registry Repository
	mockgen -package=organization -destination=internal/organization/service_mock.go apps/api/internal/organization Service
	mockgen -package=organization -destination=internal/organization/repository_mock.go apps/api/internal/organization Repository
	mockgen -package=email -destination=pkg/email/email_mock.go apps/api/pkg/email Service

dev:
	MODE=development go run main.go

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

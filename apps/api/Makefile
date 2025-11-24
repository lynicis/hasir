generate-mocks:
	mockgen -package=user -destination=internal/user/service_mock.go apps/api/internal/user Service
	mockgen -package=user -destination=internal/user/repository_mock.go apps/api/internal/user Repository

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

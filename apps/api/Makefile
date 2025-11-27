generate-mocks:
	mockgen -package=user -destination=internal/user/service_mock.go apps/api/internal/user Service
	mockgen -package=user -destination=internal/user/repository_mock.go apps/api/internal/user Repository
	mockgen -package=repository -destination=internal/repository/service_mock.go apps/api/internal/repository Service
	mockgen -package=repository -destination=internal/repository/repository_mock.go apps/api/internal/repository Repository

dev:
	MODE=development go run main.go

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

generate-mocks:
	mockgen -package=user -destination=internal/user/service_mock.go hasir-api/internal/user Service
	mockgen -package=user -destination=internal/user/repository_mock.go hasir-api/internal/user Repository
	mockgen -package=registry -destination=internal/registry/service_mock.go hasir-api/internal/registry Service
	mockgen -package=registry -destination=internal/registry/repository_mock.go hasir-api/internal/registry Repository
	mockgen -package=registry -destination=internal/registry/queue_mock.go hasir-api/internal/registry SdkGenerationQueue
	mockgen -package=organization -destination=internal/organization/service_mock.go hasir-api/internal/organization Service
	mockgen -package=organization -destination=internal/organization/repository_mock.go hasir-api/internal/organization Repository
	mockgen -package=organization -destination=internal/organization/queue_mock.go hasir-api/internal/organization Queue
	mockgen -package=email -destination=pkg/email/email_mock.go hasir-api/pkg/email Service
	mockgen -package=authorization -destination=pkg/authorization/authorization_mock.go hasir-api/pkg/authorization MemberRoleChecker

dev:
	MODE=development go run main.go

sec:
	gosec -exclude-dir=sdk -exclude-dir=repos ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

run-postgres:
	docker run -p 5432:5432 --name postgres -d -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test postgres:alpine

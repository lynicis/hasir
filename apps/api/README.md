<div align="center">
  <img src="https://github.com/protohasir/dashboard/blob/master/public/logo.webp" alt="Hasir Dashboard Logo" width="150">

# Hasir API

**A self-hosted protobuf schema registry backend service built with Go and Connect-RPC**

[![CI](https://github.com/protohasir/api/actions/workflows/ci.yaml/badge.svg)](https://github.com/protohasir/api/actions)
[![codecov](https://codecov.io/gh/protohasir/api/graph/badge.svg?token=1772BU1JL0)](https://codecov.io/gh/protohasir/api)
[![CodeQL](https://github.com/protohasir/api/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/protohasir/api/actions/workflows/github-code-scanning/codeql)
[![Go Version](https://img.shields.io/badge/Go-1.25.4-00ADD8?logo=go)](https://go.dev/)
[![Connect-RPC](https://img.shields.io/badge/Connect--RPC-Protocol-5C4EE5)](https://connectrpc.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-336791?logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
</div>

Hasir API is a protobuf schema registry backend service that provides user management, organization handling, and Git repository management through a modern gRPC-compatible API. Built with Connect-RPC over HTTP/2, it offers high performance and type-safe communication using Protocol Buffers.

## Features

- **User Management**: Registration, authentication, and profile management with JWT-based auth
- **Organization Support**: Create and manage organizations with member invitations
- **Git Repository Management**: Clone, store, and manage Git repositories
- **Connect-RPC API**: gRPC-compatible protocol over HTTP/2 with Protocol Buffers
- **Email Notifications**: SMTP-based email service for organization invitations
- **Background Job Processing**: Asynchronous email job processing
- **OpenTelemetry Integration**: Built-in distributed tracing support
- **Database Migrations**: Automated schema migrations on startup

## Architecture

Hasir API follows a domain-driven design with three core domains:

```text
internal/
├── user/           # User registration, authentication, and management
├── organization/   # Organization CRUD and member management
└── registry/       # Git repository management
```

Each domain implements a layered architecture:

- **Handler**: Connect-RPC handlers (HTTP layer)
- **Service**: Business logic layer
- **Repository**: Data access layer (PostgreSQL)

### Key Technologies

- **Go 1.25.4**: Core programming language
- **Connect-RPC**: Modern RPC framework compatible with gRPC
- **PostgreSQL**: Primary data store
- **Protocol Buffers**: API definitions and serialization
- **JWT**: Authentication tokens
- **testcontainers-go**: Integration testing
- **OpenTelemetry**: Observability and tracing

## Getting Started

### Prerequisites

- Go 1.25.4 or higher
- PostgreSQL 14+
- Make (optional, for convenience commands)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/hasir-api.git
   cd hasir-api
   ```

1. Install dependencies:

   ```bash
   go mod download
   ```

1. Set up configuration:

   ```bash
   cp config.example.json config.json
   # Edit config.json with your settings
   ```

1. Run database migrations (automatic on startup)

### Configuration

Hasir API supports two configuration modes:

#### Development Mode

Set `MODE=development` and use `config.json`:

```json
{
  "postgresql": {
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "postgres",
    "database": "hasir"
  },
  "jwt": {
    "secret": "your-secret-key",
    "accessTokenExpiration": "15m",
    "refreshTokenExpiration": "7d"
  }
}
```

#### Production Mode

Use environment variables with `HASIR_` prefix:

```bash
export HASIR_POSTGRESQL_HOST=localhost
export HASIR_POSTGRESQL_PORT=5432
export HASIR_JWT_SECRET=your-secret-key
```

See [config.example.json](config.example.json) for all available options.

## Usage

### Running the Server

Development mode:

```bash
make dev
```

Production mode:

```bash
go run main.go
```

The server will start on the configured port (default: 8080) and automatically apply database migrations.

### Testing

Run all tests:

```bash
go test ./...
```

Run tests for a specific package:

```bash
go test ./internal/user
```

Run a specific test:

```bash
go test ./internal/user -run TestServiceRegister
```

### Code Quality

Run linter:

```bash
make lint
```

Auto-fix linting issues:

```bash
make lint-fix
```

Generate mocks (after modifying interfaces):

```bash
make generate-mocks
```

## API Documentation

The API uses Connect-RPC protocol with Protocol Buffers. API definitions are available at [buf.build/hasir/hasir](https://buf.build/hasir/hasir).

### Authentication

Most endpoints require JWT authentication. Include the access token in the `Authorization` header:

```text
Authorization: Bearer <access_token>
```

Public endpoints (no authentication required):

- User registration
- User login
- Token renewal

### Example: User Registration

```bash
curl -X POST http://localhost:8080/hasir.user.v1.UserService/Register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

## Project Structure

```text
hasir-api/
├── cmd/                    # Application entrypoints
├── internal/               # Private application code
│   ├── user/              # User domain
│   ├── organization/      # Organization domain
│   ├── registry/          # Git registry domain
│   └── internal.go        # Shared interfaces
├── pkg/                   # Public libraries
│   ├── auth/             # JWT authentication
│   ├── config/           # Configuration management
│   ├── email/            # Email service
│   ├── organization/     # Organization utilities
│   └── proto/            # Generated Protocol Buffers
├── migrations/           # Database migrations
├── repos/               # Git repository storage
├── config.json          # Development configuration
├── main.go             # Application entry point
└── Makefile            # Build and development commands
```

## Development

### Adding a New RPC Method

1. Update Protocol Buffer definitions at buf.build/hasir/hasir
2. Update the handler implementation
3. Update the service layer
4. Update the repository if needed
5. Add tests for all layers

### Database Migrations

Create new migration files in `migrations/`:

```bash
# Format: {version}_{description}.up.sql and {version}_{description}.down.sql
migrations/
├── 000001_initial_schema.up.sql
├── 000001_initial_schema.down.sql
├── 000002_add_organizations.up.sql
└── 000002_add_organizations.down.sql
```

Migrations are automatically applied on server startup.

### Testing Strategy

- **Repository layer**: Integration tests with real PostgreSQL via testcontainers
- **Service layer**: Unit tests with mocked repositories
- **Handler layer**: Unit tests with mocked services

All tests follow table-driven testing patterns for comprehensive coverage.

## Deployment

### Docker

Build the Docker image:

```bash
docker build -t hasir-api .
```

Run with Docker Compose:

```bash
docker-compose up -d
```

### Environment Variables

Required environment variables for production:

- `HASIR_POSTGRESQL_HOST`: PostgreSQL host
- `HASIR_POSTGRESQL_PORT`: PostgreSQL port
- `HASIR_POSTGRESQL_USER`: Database user
- `HASIR_POSTGRESQL_PASSWORD`: Database password
- `HASIR_POSTGRESQL_DATABASE`: Database name
- `HASIR_JWT_SECRET`: JWT signing secret

See [config.example.json](config.example.json) for all configuration options.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Pre-commit Hooks

The repository uses pre-commit hooks that run:

- `gofmt` and `goimports` for formatting
- `go test ./...` for testing
- `golangci-lint` for linting

Ensure all hooks pass before submitting a PR.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Connect-RPC](https://connectrpc.com/) for the excellent RPC framework
- [Protocol Buffers](https://protobuf.dev/) for type-safe API definitions
- [testcontainers-go](https://golang.testcontainers.org/) for integration testing
- [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations

## Support

For support, please open an issue in the [GitHub issue tracker](https://github.com/yourusername/hasir-api/issues).

---

<div align="center">
  Made with ❤️ by the Hasir team
</div>

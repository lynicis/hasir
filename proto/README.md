<div align="center">
  <img src="https://github.com/protohasir/dashboard/blob/master/public/logo.webp" alt="Hasir Dashboard Logo" width="150">

# Hasir Proto

**Protocol Buffer definitions for the Hasir platform - a collaborative protobuf registry and code generation service**

[![Publish Schema to BSR](https://github.com/protohasir/proto/actions/workflows/master.yaml/badge.svg)](https://github.com/protohasir/proto/actions)
[![Buf](https://img.shields.io/badge/Buf-Build-blue)](https://buf.build/hasir/hasir)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
</div>

## Table of Contents

- [Hasir Proto](#hasir-proto)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Linting](#linting)
    - [Formatting](#formatting)
    - [Generating Code](#generating-code)
    - [Breaking Change Detection](#breaking-change-detection)
  - [API Services](#api-services)
    - [UserService](#userservice)
    - [OrganizationService](#organizationservice)
    - [RegistryService](#registryservice)
  - [Project Structure](#project-structure)
  - [Development](#development)
    - [Adding New Services](#adding-new-services)
    - [Versioning](#versioning)
    - [Dependencies](#dependencies)
  - [Code Quality](#code-quality)
    - [Pre-commit Hooks](#pre-commit-hooks)
    - [Linting Rules](#linting-rules)
    - [Breaking Change Detection](#breaking-change-detection-1)
  - [Contributing](#contributing)
  - [License](#license)

## Overview

Hasir Proto contains the Protocol Buffer definitions for the Hasir platform, a service that enables teams to collaborate on protobuf schemas, manage organizations and repositories, and generate SDKs in multiple languages.

This repository defines three core services:
- **UserService**: Authentication, API/SSH key management
- **OrganizationService**: Team collaboration and member management
- **RegistryService**: Repository management, file browsing, and commit history

## Features

- **Type-safe API definitions** using Protocol Buffers v3
- **Built-in validation** with [buf validate](https://github.com/bufbuild/protovalidate)
- **Automated linting** with Buf's BASIC style guide
- **Breaking change detection** at the file level
- **Pre-commit hooks** for formatting and validation
- **Multi-language SDK support** (Go, JavaScript, and more)

## Prerequisites

- [Buf CLI](https://buf.build/docs/installation) - Protocol buffer tooling
- [Git](https://git-scm.com/) - Version control
- [pre-commit](https://pre-commit.com/) (optional) - Git hook management

## Installation

1. Clone the repository:
```bash
git clone https://github.com/protohasir/hasir-proto.git
cd hasir-proto
```

2. Install Buf CLI:
```bash
# macOS
brew install bufbuild/buf/buf

# Other platforms
# See https://buf.build/docs/installation
```

3. (Optional) Set up pre-commit hooks:
```bash
pre-commit install
```

## Usage

### Linting

Check your protobuf definitions for style violations:

```bash
make lint
# or
buf lint
```

### Formatting

Auto-format all `.proto` files:

```bash
make fix-formatting
# or
buf format --write
```

### Generating Code

The repository is configured to generate Go code with the package prefix `github.com/protohasir/proto/gen/go`. Code generation is managed via [buf.gen.yaml](buf.gen.yaml).

To generate code:

```bash
buf generate
```

### Breaking Change Detection

Buf will automatically check for breaking changes against the previous version:

```bash
buf breaking --against '.git#branch=main'
```

## API Services

### UserService

Handles user authentication and credential management.

**Endpoints:**
- `Register` - Create a new user account
- `Login` - Authenticate and receive tokens
- `RenewTokens` - Refresh access tokens
- `ForgotPassword` - Initiate password reset
- `ResetPassword` - Complete password reset
- `UpdateUser` - Modify user profile
- `DeleteAccount` - Permanently delete account
- `CreateApiKey` / `GetApiKeys` / `RevokeApiKey` - API key management
- `CreateSshKey` / `GetSshKeys` / `RevokeSshKey` - SSH key management

See [proto/user/v1/user.proto](proto/user/v1/user.proto) for details.

### OrganizationService

Manages organizations and team collaboration.

**Endpoints:**
- `CreateOrganization` - Create a new organization
- `GetOrganizations` / `GetOrganization` - List and retrieve organizations
- `UpdateOrganization` / `DeleteOrganization` - Modify organization settings
- `InviteMember` - Send organization invitations
- `IsInvitationValid` / `RespondToInvitation` - Handle invitations
- `GetMembers` - List organization members
- `UpdateMemberRole` - Change member permissions
- `DeleteMember` - Remove members

Organizations support role-based access control (Reader, Author, Owner) and public/private visibility.

See [proto/organization/v1/organization.proto](proto/organization/v1/organization.proto) for details.

### RegistryService

Manages protobuf repositories, commits, and file browsing.

**Endpoints:**
- `CreateRepository` / `GetRepositories` / `GetRepository` - Repository CRUD operations
- `UpdateRepository` / `DeleteRepository` - Modify repositories
- `GetCommits` - View commit history
- `GetFileTree` - Browse repository file structure
- `GetFilePreview` - View file contents
- `UpdateSdkPreferences` - Configure SDK generation preferences

Supports multiple SDK targets including Go (protobuf, gRPC, ConnectRPC) and JavaScript (Bufbuild ES, protobuf, ConnectRPC).

See [proto/registry/v1/registry.proto](proto/registry/v1/registry.proto) for details.

## Project Structure

```
hasir-proto/
├── proto/
│   ├── user/v1/              # User authentication service
│   │   └── user.proto
│   ├── organization/v1/      # Organization management service
│   │   └── organization.proto
│   ├── registry/v1/          # Repository and file management service
│   │   └── registry.proto
│   └── shared/               # Shared types and enums
│       ├── filter.proto      # Pagination utilities
│       ├── role.proto        # Access control roles
│       └── visibility.proto  # Public/private visibility
├── buf.yaml                  # Buf configuration
├── buf.gen.yaml              # Code generation config
├── buf.lock                  # Dependency lock file
├── Makefile                  # Build automation
└── .pre-commit-config.yaml   # Git hooks configuration
```

## Development

### Adding New Services

1. Create a new directory under `proto/` following the pattern `{service}/v1/`
2. Define your service in a `.proto` file
3. Update [buf.yaml](buf.yaml) if needed
4. Run `make lint` to validate
5. Run `make fix-formatting` to format

### Versioning

This project follows semantic versioning for protobuf APIs:
- Services are versioned (e.g., `user.v1`, `organization.v1`)
- Breaking changes require a new major version (e.g., `v2`)
- The Buf breaking change detector enforces file-level compatibility

### Dependencies

External dependencies are managed via Buf Schema Registry:
- `buf.build/protocolbuffers/wellknowntypes` - Standard protobuf types
- `buf.build/bufbuild/protovalidate` - Field validation annotations

## Code Quality

### Pre-commit Hooks

This repository uses pre-commit hooks to maintain code quality:

- **trailing-whitespace**: Remove trailing spaces
- **end-of-file-fixer**: Ensure files end with a newline
- **check-yaml**: Validate YAML syntax
- **buf-format**: Auto-format `.proto` files
- **buf-lint**: Lint protobuf definitions

### Linting Rules

Follows Buf's [BASIC](https://buf.build/docs/lint/rules#basic) lint rules:
- Package names match directory structure
- Service and message naming conventions
- Field naming consistency
- Proper use of reserved keywords

### Breaking Change Detection

Uses Buf's [FILE](https://buf.build/docs/breaking/rules#file) breaking change rules to prevent:
- Removing or renaming fields
- Changing field types
- Removing services or methods
- Other backward-incompatible changes

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run `make lint` and `make fix-formatting`
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

Please ensure:
- All proto files are properly formatted
- Linting passes without errors
- No breaking changes are introduced (unless intentional and documented)
- Commit messages are clear and descriptive

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">
  Made with ❤️ by the Hasir team
</div>

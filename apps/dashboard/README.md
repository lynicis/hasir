<div align="center">
  <img src="public/logo.webp" alt="Hasir Dashboard Logo" width="150">

  # Hasir Dashboard

  **Web interface for on-premise proto schema registry**

  [![CI](https://github.com/protohasir/dashboard/actions/workflows/ci.yaml/badge.svg)](https://github.com/protohasir/dashboard/actions)
  [![codecov](https://codecov.io/gh/protohasir/dashboard/branch/master/graph/badge.svg?token=3WH38YDL1T)](https://codecov.io/gh/protohasir/dashboard)
  [![Next.js](https://img.shields.io/badge/Next.js-16-black)](https://nextjs.org/)
  [![React](https://img.shields.io/badge/React-19-blue)](https://react.dev/)
  [![TypeScript](https://img.shields.io/badge/TypeScript-5-blue)](https://www.typescriptlang.org/)
  [![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
</div>

---

## Table of Contents

- [Hasir Dashboard](#hasir-dashboard)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Tech Stack](#tech-stack)
    - [Core](#core)
    - [Data \& Communication](#data--communication)
    - [UI \& Styling](#ui--styling)
    - [Forms \& Validation](#forms--validation)
    - [Development Tools](#development-tools)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Environment Variables](#environment-variables)
    - [Running the Application](#running-the-application)
  - [Development](#development)
    - [Project Structure](#project-structure)
    - [Code Quality](#code-quality)
    - [Testing](#testing)
  - [Deployment](#deployment)
    - [Docker](#docker)
    - [Production Build](#production-build)
  - [Architecture](#architecture)
    - [Transport Layer](#transport-layer)
    - [Authentication](#authentication)
    - [State Management](#state-management)
    - [Routing](#routing)
  - [Contributing](#contributing)
    - [Commit Convention](#commit-convention)
  - [License](#license)

---

## Overview

Hasir Dashboard is a modern, full-featured web application built with Next.js 16 that serves as the frontend interface for an on-premise protobuf schema registry. It provides a seamless user experience for managing and browsing protocol buffer schemas with real-time updates via gRPC-web.

The application leverages cutting-edge React patterns including server components, streaming, and the latest Next.js App Router features to deliver optimal performance and developer experience.

## Features

- **Schema Management**: Browse, search, and manage protobuf schemas
- **Real-time Updates**: Live schema updates via gRPC-web transport
- **Authentication**: Secure JWT-based authentication with refresh tokens
- **Dark Mode**: Full dark/light theme support
- **Responsive Design**: Mobile-first responsive UI using Tailwind CSS
- **Type-Safe**: End-to-end type safety with TypeScript and generated protobuf types
- **Modern UI**: Built with shadcn/ui components and Radix UI primitives
- **Fast Development**: Turbopack-powered dev server for instant feedback
- **Comprehensive Testing**: Full test coverage with Vitest and React Testing Library

## Tech Stack

### Core
- [Next.js 16](https://nextjs.org/) - React framework with App Router
- [React 19](https://react.dev/) - UI library with latest features
- [TypeScript 5](https://www.typescriptlang.org/) - Type-safe JavaScript

### Data & Communication
- [Connect-RPC](https://connectrpc.com/) - gRPC-web client
- [TanStack Query](https://tanstack.com/query) - Server state management
- [Zustand](https://zustand-demo.pmnd.rs/) - Client state management

### UI & Styling
- [Tailwind CSS 4](https://tailwindcss.com/) - Utility-first CSS
- [shadcn/ui](https://ui.shadcn.com/) - Re-usable component library
- [Radix UI](https://www.radix-ui.com/) - Accessible UI primitives
- [Lucide Icons](https://lucide.dev/) - Icon library
- [Framer Motion](https://www.framer.com/motion/) - Animation library

### Forms & Validation
- [React Hook Form](https://react-hook-form.com/) - Performant forms
- [Zod](https://zod.dev/) - Schema validation

### Development Tools
- [Bun](https://bun.sh/) - Fast JavaScript runtime
- [Vitest](https://vitest.dev/) - Unit testing framework
- [ESLint](https://eslint.org/) - Code linting
- [Turbopack](https://turbo.build/pack) - Fast bundler

## Getting Started

### Prerequisites

- [Bun](https://bun.sh/) >= 1.0 (or Node.js >= 18)
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/lynicis/hasir-dashboard.git
cd hasir-dashboard
```

2. Install dependencies:
```bash
bun install
```

### Environment Variables

Create a `.env` file in the root directory:

```bash
cp .env.example .env
```

Configure the following environment variables:

```env
# Backend API URL (gRPC-web endpoint)
NEXT_PUBLIC_API_URL=http://localhost:8080

# Frontend base URL
NEXT_PUBLIC_BASE_URL=http://localhost:3000
```

### Running the Application

Start the development server:

```bash
bun dev
```

The application will be available at [http://localhost:3000](http://localhost:3000).

## Development

### Project Structure

```
hasir-dashboard/
├── app/                    # Next.js App Router pages and layouts
│   ├── (authenticated)/    # Protected routes group
│   ├── login/             # Login page
│   └── providers.tsx      # App-level providers
├── components/            # React components
│   ├── ui/               # shadcn/ui components
│   └── *.tsx             # Feature components
├── lib/                   # Utility functions and helpers
│   ├── use-client.ts     # Connect-RPC client hook
│   └── utils.ts          # Common utilities
├── stores/               # Zustand state stores
│   └── user-store.ts     # User authentication store
├── public/               # Static assets
└── tests/                # Test utilities and setup
```

### Code Quality

Run the linter:

```bash
# Check for issues
bun run lint

# Auto-fix issues
bun run lint:fix
```

The project uses:
- **ESLint** with Next.js recommended config
- **Perfectionist** plugin for consistent import/export ordering
- **TypeScript** strict mode for type safety

### Testing

Run tests:

```bash
# Watch mode
bun test

# CI mode with coverage
bun run test:ci
```

Tests are written using:
- **Vitest** for test running
- **React Testing Library** for component testing
- **@testing-library/user-event** for user interaction simulation

All components should have corresponding `.test.tsx` files with comprehensive test coverage.

## Deployment

### Docker

Build and run with Docker:

```bash
# Build image
docker build -t hasir-dashboard .

# Run container
docker run -p 3000:3000 -e NEXT_PUBLIC_API_URL=http://your-api:8080 hasir-dashboard
```

### Production Build

Create an optimized production build:

```bash
# Build the application
bun run build

# Start production server
bun start
```

The build output is configured for standalone mode, making it ideal for containerized deployments.

## Architecture

### Transport Layer

The application uses **Connect-RPC** (gRPC-web) for backend communication:

- Transport configured in [app/providers.tsx](app/providers.tsx)
- Binary wire format for optimal performance
- Custom `useClient` hook in [lib/use-client.ts](lib/use-client.ts)
- Generated protobuf types from `@buf/hasir_hasir.bufbuild_es`

```typescript
// Example usage
import { useClient } from '@/lib/use-client';
import { SchemaService } from '@buf/hasir_hasir.connectrpc_es/schema/v1/service_connect';

const client = useClient(SchemaService);
```

### Authentication

JWT-based authentication with the following features:

- Access and refresh token pair
- Tokens stored in HTTP-only cookies (7-day expiration)
- Automatic token refresh on expiration
- User state managed via Zustand store
- Protected routes using route groups

Authentication flow:
1. User logs in via `/login`
2. Server returns JWT tokens stored in cookies
3. Tokens included automatically in subsequent requests
4. Refresh token used to obtain new access token when expired

### State Management

**Global State (Zustand)**:
- User authentication state
- User profile data
- Session management

**Server State (TanStack Query)**:
- Data fetching and caching
- Optimistic updates
- Background refetching
- Request deduplication

**UI State**:
- Local component state using React hooks
- Form state via React Hook Form
- Theme preference via next-themes

### Routing

The application uses Next.js App Router with route groups:

**Public Routes**:
- `/login` - User login
- `/register` - User registration
- `/invite/[token]` - Accept invitation

**Authenticated Routes** (`/(authenticated)/`):
- `/dashboard` - Main dashboard
- `/profile` - User profile management

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting (`bun test && bun run lint`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Commit Convention

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Test additions or changes
- `chore:` - Build process or auxiliary tool changes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">
  Made with ❤️ by the Hasir team
</div>

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hasir Dashboard is a Next.js 16 application serving as a web interface for an on-premise proto schema registry. It uses the App Router with TypeScript and is built with modern React patterns including server components.

## Development Commands

- **Development server**: `bun dev` - starts Next.js development server with Turbopack
- **Build**: `bun run build` - creates production build
- **Production server**: `bun start` - runs production server
- **Linting**: `bun run lint` - runs ESLint, `bun run lint:fix` - auto-fixes issues
- **Testing**: `bun run test` - runs Vitest in watch mode, `bun run test:ci` - runs tests with coverage for CI

## Architecture

### Core Technologies
- **Next.js 16** with App Router and React 19
- **Connect-RPC** for gRPC-web communication with backend services
- **Zustand** for client-side state management
- **TanStack Query** for server state management
- **shadcn/ui** components with Tailwind CSS
- **Vitest** for testing with React Testing Library

### Transport Layer
The application uses Connect-RPC (gRPC-web) to communicate with backend services:
- Main transport configured in `app/providers.tsx` with binary format
- Environment variable `NEXT_PUBLIC_API_URL` sets the backend base URL
- Custom `useClient` hook in `lib/use-client.ts` for service clients
- Generated protobuf types from `@buf/hasir_hasir.bufbuild_es`

### Authentication
- JWT-based authentication with access/refresh tokens
- Tokens stored in cookies (7-day expiration by default)
- User state managed via Zustand store (`stores/user-store.ts`)
- Authentication context provided through `UserStoreProvider`

### Routing Structure
- Public routes: `/login`, `/register`
- Authenticated routes under `/(authenticated)/` group:
  - `/dashboard` - main dashboard
  - `/profile` - user profile management
  - `/invite/[token]` - invitation acceptance

### State Management
- **Global state**: Zustand store for user authentication state
- **Server state**: TanStack Query for data fetching and caching
- **UI state**: Local component state and React hooks

### Styling
- Tailwind CSS 4 with CSS variables for theming
- Dark/light mode support via `next-themes`
- shadcn/ui component library with "new-york" style variant
- Custom CSS in `app/globals.css`

## Key Patterns

### Component Organization
- UI components in `components/ui/` (from shadcn/ui)
- Feature components in `components/` root
- Each component has corresponding `.test.tsx` file
- Forms use `react-hook-form` with Zod validation

### Code Style
- ESLint with Next.js config and perfectionist plugin for import/export sorting
- TypeScript strict mode enabled
- Path aliases: `@/` maps to project root

### Testing
- Vitest with jsdom environment
- React Testing Library for component testing
- Test setup in `vitest.setup.ts`
- Coverage reports with @vitest/coverage-v8

### Environment Variables
- `NEXT_PUBLIC_API_URL` - Backend API base URL
- `NEXT_PUBLIC_BASE_URL` - Frontend base URL (used in useClient hook)

## Build & Deployment
- Standalone output mode for containerization
- Dockerfile provided for production deployment
- Turbopack enabled for faster development builds
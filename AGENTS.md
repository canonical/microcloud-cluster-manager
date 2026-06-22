# Microcloud Cluster Manager Agent Guidelines

This document provides guidance for AI agents working with the Microcloud Cluster Manager codebase.

## Project Overview

**Microcloud Cluster Manager** is a management system for MicroCloud cluster orchestration. It provides both a REST API backend and a web-based frontend for managing distributed cloud clusters.

- **Backend Language**: Go
- **Frontend Language**: TypeScript + React
- **Frontend Build Tool**: Vite
- **Frontend Package Manager**: Yarn
- **Testing**: Go testing package (backend), Vitest (frontend unit), Playwright (frontend E2E)
- **Code Quality**: golangci-lint (Go), ESLint/Prettier (TypeScript)
- **License**: AGPL-3.0-only (root project), LGPL-3.0-only (UI)

## Technology Stack

### Backend (Go)

- **Go**: Core backend language
- **HTTP Router**: Gorilla Mux (`github.com/gorilla/mux`) on top of `net/http`
- **Database**: PostgreSQL with sqlx and Goose migrations (see internal/pkg/database)
- **Configuration Management**: Environment-variable-based config (internal/pkg/config)
- **Logging**: Custom logger in internal/pkg/logger

### Frontend (TypeScript/React)

- **React**: UI framework
- **React Router DOM**: Client-side routing
- **TanStack React Query**: Data fetching and caching
- **Formik**: Form state management
- **Canonical React Components**: Canonical design system
- **Fetch API**: Uses `fetch` for API communication (see `ui/src/api/*.ts`)
- **TypeScript**: Type-safe JavaScript

## Project Structure

```
microcloud-cluster-manager/
+-- cmd/                                 # Command-line entry points
|   +-- main.go                          # Main application entry
|   +-- cli/
|   |   +-- cli.go                       # CLI interface
|   |   +-- cli_enroll.go                # CLI enroll command
|   +-- cluster-connector/
|   |   +-- cluster-connector.go         # Cluster connector service
|   +-- management-api/
|   |   +-- management-api.go            # Management API server
+-- internal/                            # Private Go packages
|   +-- app/
|   |   +-- cluster-connector/
|   |   |   +-- api/                     # Cluster connector API handlers
|   |   |   +-- core/                    # Cluster connector business logic
|   |   +-- management-api/
|   |   |   +-- api/                     # Management API handlers
|   |   |   +-- core/                    # Management API business logic
|   +-- pkg/
|   |   +-- api/
|   |   |   +-- api.go                   # API utilities
|   |   |   +-- rest.go                  # REST API helpers
|   |   |   +-- models/                  # API response models
|   |   +-- config/                      # Configuration parsing
|   |   |   +-- cipher.go                # Encryption utilities
|   |   +-- database/                    # Database layer
|   |   +-- logger/                      # Logging utilities
|   |   +-- middleware/                  # HTTP middleware
|   |   +-- request/                     # Request utilities
|   |   +-- types/                       # Common type definitions
+-- ui/                                  # Frontend (React/TypeScript)
|   +-- src/
|   |   +-- App.tsx                      # App routing and layout
|   |   +-- Root.tsx                     # Root component with providers
|   |   +-- index.tsx                    # Entry point
|   |   +-- api/                         # API calls and queries
|   |   |   +-- *.ts                     # Resource-specific API files
|   |   +-- components/                  # Reusable React components
|   |   |   +-- forms/                   # Form components
|   |   +-- context/                     # React Context and custom hooks
|   |   |   +-- appProviders.tsx         # Main app context provider setup
|   |   |   +-- auth.tsx                 # Authentication context
|   |   |   +-- use*.tsx                 # Custom hooks
|   |   +-- pages/                       # Page components by feature
|   |   +-- sass/                        # SCSS stylesheets
|   |   +-- types/                       # TypeScript type definitions
|   |   +-- util/                        # Utility functions
|   +-- tests/                           # End-to-end tests
|   |   +-- *.spec.ts                    # Playwright test files
|   |   +-- fixtures/                    # Test fixtures and setup
|   |   +-- helpers/                     # Test helper functions
|   +-- public/                          # Static assets
|   +-- package.json                     # Dependencies
|   +-- vite.config.ts                   # Vite build configuration
|   +-- vitest.config.ts                 # Unit test configuration
|   +-- playwright.config.ts             # E2E test configuration
|   +-- tsconfig.json                    # TypeScript configuration
+-- test/                                # Backend tests
|   +-- e2e/                             # End-to-end tests
|   |   +-- *_test.go                    # Go E2E test files
|   |   +-- e2e_test.go                  # Main E2E test setup
|   |   +-- helpers/                     # Test helpers
|   +-- unit/                            # Unit tests
|   |   +-- *_test.go                    # Go unit test files
|   +-- helpers/                         # Shared test utilities
|   +-- keys/                            # Test certificates and keys
+-- scripts/                             # Build and setup scripts
|   +-- run-backend.sh                   # Set up environment and start backend server
|   +-- run-coverage.sh                  # Run coverage tests
|   +-- setup-env.sh                     # Environment setup
|   +-- pre-commit.sh                    # Pre-commit hooks
+-- go.mod                               # Go module definition
+-- makefile                             # Build targets
+-- ARCHITECTURE.md                      # Architecture documentation
+-- CONTRIBUTING.md                      # Contribution guidelines
+-- README.md                            # Project overview
+-- SECURITY.md                          # Security guidelines
```

## Backend (Go) Structure

### `/cmd` - Entry Points

- **`main.go`**: Application bootstrap
- **`cli/cli.go`**: Command-line interface definitions
- **`cluster-connector/cluster-connector.go`**: Cluster connector service
- **`management-api/management-api.go`**: Management API server

### `/internal/app` - Application Logic

**Cluster Connector** (`cluster-connector/`):

- **`api/`**: HTTP handlers for cluster connector endpoints
- **`core/`**: Business logic for cluster connection management

**Management API** (`management-api/`):

- **`api/`**: HTTP handlers for management API endpoints
- **`core/`**: Business logic for cluster management

### `/internal/pkg` - Shared Packages

- **`api/`**: REST API utilities and response models
- **`config/`**: Configuration parsing and encryption (see `cipher.go` for secrets handling)
- **`database/`**: Data persistence layer
- **`logger/`**: Structured logging with request context
- **`middleware/`**: HTTP middleware (request tracing/logging)
- **`request/`**: Request handling and validation utilities
- **`types/`**: Shared type definitions including `Authorizor` interface and `UserInfo`

## Frontend (TypeScript/React) Structure

### `/ui/src/api`

- **Purpose**: Contains all API calls to the backend
- **Pattern**: Each resource has a dedicated file (clusters.ts, tokens.ts, etc.)
- **Usage**: Wrapped with React Query for caching and state management
- **Pattern**: Files contain functions like `fetchClusters()`, `createCluster()`, `deleteCluster()`

### `/ui/src/components`

- **Purpose**: Reusable UI components following Canonical design system
- **Subdirectory `/ui/src/components/forms`**: Form components for data entry
- **Pattern**: Each component is a .tsx file with optional .scss file for styling

### `/ui/src/context`

- **Purpose**: React Context for global state and custom hooks
- **Naming Convention**: Hooks are named `use*` (e.g., `useClusters`, `useRemoteClusters`)
- **Query Integration**: Most hooks integrate with React Query for server state

### `/ui/src/pages`

- **Purpose**: Page/feature components organized by resource type
- **Structure**: Each feature has a subdirectory with multiple views (list, detail, form, etc.)

### `/ui/src/types`

- **Purpose**: Shared TypeScript type definitions
- **Pattern**: One file per resource type (e.g., `cluster.d.ts`, `token.d.ts`)
- **API Contract**: Types reflect the backend API response structure

### `/ui/src/util`

- **Purpose**: Pure utility functions
- **Examples**: Formatting, validation, data transformation
- **Testing**: These functions should have unit tests

## Development Setup

### Prerequisites

1. **Go**: Version specified in go.mod
2. **Node.js/Yarn**: For frontend development
3. **Make**: For build targets

### Initial Setup

**Automated Dependency Installation** (recommended for Ubuntu Linux):

```bash
# 1. Install core dependencies
make install-core

# 2. Install additional dependencies
make install-deps
```

**Manual Dependency Installation:**

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed manual installation of Go, Docker, Kubernetes (kubectl, Kind), Skaffold, Node.js (NVM), Dotrun, and Juju.

### Running Development Server

#### Backend

Run the backend services locally (Docker for dependencies, services started by `scripts/run-backend.sh`):

```bash
# Starts Postgres/Prometheus in Docker and runs the Go services locally
make dev
```

When `make dev` completes, it will have built `cmd/app`, started required dependencies (Postgres/Prometheus) in Docker, and the service logs will be printed to your terminal.

Backend services run on (default `make dev`):

- **Management API**: `https://ma.lxd-cm.local:30000`
- **Cluster Connector**: `https://cc.lxd-cm.local:9000`
- **Database**: `127.0.0.1:5432`

**Note**: First-time startup may take several minutes while resources and images are pulled.

#### Frontend

In a separate terminal, start the frontend development server:

```bash
# 1. Add local development hosts to /etc/hosts
sudo make add-hosts

# 2. Start the UI development server
make ui
```

Frontend UI runs on: `https://ma.lxd-cm.local:8414`

#### Cleanup

After finishing development for the day, clean up unused Docker images to prevent disk space issues:

```bash
make nuke
```

## Build and Run Commands

### Backend (Go)

```bash
# Build binary
go build -o microcloud-cluster-manager ./cmd

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linter
golangci-lint run ./...

# Format code
go fmt ./...
```

### Frontend (TypeScript/React)

```bash
# Start development server
make ui

# or run directly (requires PORT to be set; see ui/.env)
cd ui
PORT=8414 yarn start

# Build for production
yarn build

# Run linting
yarn lint-js

# Format code
yarn format-js

# Run unit tests
yarn test-unit

# Run E2E tests
yarn test-e2e-coverage
```

### Development Cluster Management

```bash
# Start entire backend development cluster
make dev

# Run backend E2E tests
make test-e2e

# Run frontend E2E tests
make test-ui-e2e

# Clean up unused Docker images
make nuke

# Add hosts to /etc/hosts for local development
sudo make add-hosts

# Enable pre-commit hooks
make add-hooks
```

## Testing

### Backend Tests

#### Unit and Integration Tests

```bash
# Run backend unit tests (recommended)
make test-unit

# Or run all tests including integration tests
go test ./...

# Run tests in specific package
go test ./internal/pkg/config

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Location**: `test/unit/` (see `test/unit/unit_test.go`)  
**Framework**: Go's standard testing package (with local helpers in `test/helpers/`)

#### E2E Tests

Run the backend against the development cluster:

```bash
# Run backend E2E tests
make test-e2e
```

**Location**: `test/e2e/`  
**Framework**: Go's standard testing package with helpers

### Frontend Tests

#### Unit and Integration Tests

```bash
cd ui

# Run unit tests once
yarn test-unit

# Generate coverage report
yarn test-unit-coverage
```

**Location**: Tests colocated with source code as `.spec.ts` files
**Framework**: Vitest with Jest-like syntax

#### E2E Tests

```bash
cd ui

# Step 1: Install Playwright (first time only)
npx playwright install

# Step 2: Create .env.local with OIDC credentials
cat > .env.local << EOF
OIDC_USER="cluster-manager-e2e-tests@example.org"
OIDC_PASSWORD="cluster-manager-e2e-password"
EOF

# Step 3: Run E2E tests
npx playwright test

# Or run specific test file
npx playwright test tests/clusters.spec.ts

# Run with specific browser
npx playwright test --project chromium
```

Alternatively, from the root directory:

```bash
make test-ui-e2e
```

**Location**: `ui/tests/*.spec.ts`  
**Framework**: Playwright

## Error Handling Patterns

### Backend (Go)

```go
// Return error from handler - API layer converts to HTTP response
if err != nil {
    logger.Log.Errorw("operation failed", "error", err, "resource_id", id)
    return fmt.Errorf("failed to fetch cluster: %w", err)
}

// Use consistent error response format
response.ErrorResponse(http.StatusNotFound, "Cluster not found")
response.ErrorResponse(http.StatusBadRequest, "Invalid request parameters")
response.ErrorResponse(http.StatusForbidden, "Permission denied")  // From authorization check
```

## Code Conventions

### Go

- **Package Organization**: Group related functionality in packages under `internal/`
- **Error Handling**: Always check and handle errors; wrap errors with context using `%w` format verb
- **Logging**: Use `logger.Log.Errorw()`, `logger.Log.Infow()` with structured fields
- **Naming**: Follow Go naming conventions (PascalCase for exported, camelCase for unexported)
- **Comments**: Add export comments on all public functions, types, and interfaces
- **Interfaces**: Keep interfaces small and focused (e.g., `Authorizor`, `Authenticator`)
- **Concurrency**: Use goroutines and channels for concurrent operations; handle context cancellation
- **Testing**: Include `*_test.go` files in the appropriate subdirectories under `test/`

### TypeScript/React

- **Type Annotations**: Always explicitly type component props and function returns; avoid implicit `any`
- **No `any` Type**: Avoid `any` unless absolutely necessary; use generics instead
- **Functional Components**: Use only function components with hooks
- **Props Interface**: Define props as `interface Props { ... }` for local components or in `ui/src/types/` for shared types
- **Hooks**: Use React hooks for state (`useState`), effects (`useEffect`), and context (`useContext`)
- **Custom Hooks**: Prefix with `use` (e.g., `useClusters`, `useAuth`) and place in `ui/src/context/`
- **File Naming**: PascalCase for React components (`ClusterDetail.tsx`), camelCase for utilities (`formatDate.ts`)
- **Path Aliases**: Use TypeScript paths (configured in `ui/tsconfig.json`) for clean imports: `import { useAuth } from 'context/auth'`
- **No Direct DOM Access**: Use React state and refs instead of direct DOM manipulation

### Styling

- **SCSS**: Use SCSS for component-specific styles
- **BEM Convention**: Block Element Modifier for class naming
- **Canonical Components**: Prefer `@canonical/react-components` over custom styles

## API Patterns

### Backend (Go)

#### Handlers and Routes

- **HTTP Handlers**: Defined in `internal/app/*/api/` directories (e.g., `internal/app/management-api/api/v1/remote_cluster.go`)
- **Route Registration**: Endpoints defined as `types.Endpoint` structs with handler, method, path, and authorization settings
- **Models**: Response structures in `internal/pkg/api/models/`
- **REST Conventions**: Use standard HTTP methods (GET, POST, PUT, DELETE)
- **Error Responses**: Use `response.ErrorResponse(statusCode, message)` for consistent error format

#### Authentication & Authorization

**Authentication** (per service):

- **Management API**: OIDC-based (OpenID Connect) - validates JWT tokens
- **Cluster Connector**: mTLS-based (Mutual TLS) - validates client certificates

**Authorization** (per endpoint):

- Implemented via `Authorizor` interface in `internal/pkg/types/rest.go`
- Each endpoint declares `AllowUnauthorized` flag (default: false) and `AllowedEntitlements` (for future granular permissions)
- Authorization check runs in `internal/pkg/api/rest.go` (in `registerEndpoint`) after authentication, before handler execution
- Returns 403 Forbidden if authorization fails

**Management API Authorization** (group-based RBAC):

```go
type Endpoint struct {
    Handler             func(RouteConfig) EndpointHandler
    Path                string
    Method              string
    AllowUnauthorized   bool      // Set true only for public endpoints like /v1/
    AllowedEntitlements []string  // Prepared for future granular permissions
}
```

- Currently all protected endpoints require user to be in `"admins"` OIDC group
- User groups extracted from OIDC custom claim `mcm-idp-groups` (configured in OIDC provider)
- Stored in `UserInfo` struct accessible via context: `ctx.Value(types.UserInfoKey)`

**Cluster Connector Authorization**:

- Uses no-op authorizor (all authenticated clients are authorized)
- Authentication via mTLS with certificate pinning against database

For detailed authorization patterns, see [internal/app/management-api/core/auth/](internal/app/management-api/core/auth/) and authorization notes.

### Frontend (React Query Integration)

- **Query Keys**: Organized hierarchically in `ui/src/util/queryKeys.tsx` following TanStack conventions
- **Custom Hooks**: Wrap queries in `ui/src/context/use*.tsx` hooks (e.g., `useClusters`, `useRemoteClusters`)
- **Queries**: Use `useQuery` or `useQueries` for fetch operations with caching
- **Mutations**: Use `useMutation` for create/update/delete
- **Cache Invalidation**: Invalidate related query keys on mutation success to trigger refetching
- **Error Handling**: Provide user-friendly error messages; handle API errors in catch blocks
- **Loading States**: Use `isLoading`, `isPending` flags from query/mutation results

## Common Tasks

### Adding a New API Endpoint (Backend)

1. **Define Handler**: Add in `internal/app/management-api/api/` or `cluster-connector/api/`
2. **Add Route**: Register in API router/mux
3. **Implement Logic**: Place business logic in corresponding `core/` directory
4. **Add Tests**: Create `*_test.go` file with unit tests
5. **Add E2E Tests**: Add test case in `test/e2e/`

### Adding a New Frontend Feature

1. **Create Types**: Define API types in `ui/src/types/`
2. **Create API Layer**: Add fetch functions in `ui/src/api/`
3. **Create Hook**: Add React Query integration in `ui/src/context/`
4. **Create Components**: Build UI components in `ui/src/components/`
5. **Create Pages**: Add feature pages in `ui/src/pages/`
6. **Add Tests**: Include unit tests and E2E tests

### Debugging

#### Frontend (React)

1. **Browser DevTools**: Open Chrome DevTools with F12
2. **React DevTools**: Check component props and state
3. **Network Tab**: Inspect API calls
4. **Console Logs**: Add temporary logging (remove before committing)

## Important Files to Know

| File                                         | Purpose                                |
| -------------------------------------------- | -------------------------------------- |
| `cmd/main.go`                                | Application entry point                |
| `cmd/management-api/management-api.go`       | Management API server setup            |
| `cmd/cluster-connector/cluster-connector.go` | Cluster Connector server setup         |
| `internal/pkg/types/rest.go`                 | Endpoint, Authorizor, Auth interfaces  |
| `internal/pkg/types/types.go`                | UserInfo, context keys, shared types   |
| `internal/pkg/config/cipher.go`              | Configuration and encryption           |
| `internal/pkg/database/`                     | PostgreSQL layer with migrations       |
| `internal/pkg/logger/`                       | Structured logging utilities           |
| `internal/pkg/middleware/`                   | Request tracing/logging middleware     |
| `internal/app/management-api/core/auth/`     | OIDC authenticator and authorizor      |
| `internal/app/cluster-connector/core/auth/`  | mTLS authenticator and authorizor      |
| `ui/src/App.tsx`                             | Frontend routing and layout            |
| `ui/src/Root.tsx`                            | Frontend root with providers           |
| `ui/src/context/`                            | React Context and custom hooks         |
| `ui/src/api/`                                | Backend API call functions             |
| `ui/tsconfig.json`                           | Frontend TypeScript config             |
| `go.mod`                                     | Go module definition                   |
| `makefile`                                   | Build and deployment automation        |
| `ARCHITECTURE.md`                            | System architecture and design         |
| `CONTRIBUTING.md`                            | Contribution and development process   |
| `SECURITY.md`                                | Security guidelines and best practices |

## GitHub Actions and CI/CD

### Automated on Pull Requests

- **Workflow File**: `.github/workflows/` (if present)
- **Checks**:
  - Go linting with golangci-lint
  - Go tests with coverage (unit and E2E)
  - Backend binary build verification
  - Frontend linting (ESLint, Prettier)
  - Frontend type checking (TypeScript)
  - Frontend tests (Vitest unit tests)
  - Frontend E2E tests (Playwright)

## Tips for Working with Microcloud Cluster Manager

1. **Environment Variables**: Use `.env.local` to override UI defaults
2. **Backend-Frontend Communication**: Ensure API contracts are maintained between backend handlers and frontend types
3. **Database Migrations**: If applicable, manage schema changes systematically
4. **Configuration**: Centralized config handling in `internal/pkg/config/`
5. **Logging**: Use the logger utilities for consistent logging across the application
6. **Error Handling**: Maintain consistent error response formats for better client-side handling
7. **Authentication**: Review auth middleware in `internal/pkg/middleware/` for security
8. **Authorization**: Group-based RBAC enforced via `Authorizor` interface in `internal/pkg/types/rest.go`. Each endpoint declares `AllowedEntitlements` (currently uses OIDC groups like "admins"). Authorization checked in handlers via `CheckPermissions()` after authentication. Returns 403 if user lacks required group membership.
9. **Testing**: Run tests frequently during development to catch issues early
10. **Type Safety**: Keep frontend types in sync with backend API responses
11. **Performance**: Monitor database queries and API response times

## Safety Notes

- Never commit secrets, credentials, or private keys
- Keep machine-specific overrides in `.env.local` (already gitignored)
- Do not commit generated test certificates or artifacts
- Protect sensitive data in configuration and encryption utilities

## Definition of Done (Agent Changes)

1. Use real feature/resource names in paths and symbols (no placeholder values)
2. Run relevant validation before finalizing:
   - `go test ./...` for backend changes
   - `go fmt ./...` for formatting
   - `yarn lint-js` for frontend changes
   - Targeted tests for touched areas
3. Confirm commands are runnable for this repository
4. Keep changes scoped and consistent with existing patterns in:
   - `internal/app/*/api/` and `internal/app/*/core/` (backend)
   - `ui/src/api/`, `ui/src/context/`, `ui/src/pages/`, `ui/src/types/` (frontend)

## Quick Reference: Key Scripts

```bash
# Dependency Installation
make install-core                      # Install core dependencies
make install-deps                      # Install additional dependencies

# Development Server
make dev                               # Start backend development cluster
make ui                                # Start UI development server
sudo make add-hosts                    # Add local development hosts

# Backend
make test-unit                         # Run backend unit tests
go test ./...                          # Run all backend tests
go build -o microcloud-cluster-manager ./cmd  # Build binary
golangci-lint run ./...                # Lint Go code
go fmt ./...                           # Format Go code

# Frontend (from ui/ directory)
yarn start                             # Start dev server
yarn build                             # Production build
yarn lint-js                           # Lint and type check
yarn format-js                         # Format code
yarn test-unit                         # Unit tests
npx playwright test                    # E2E tests

# Testing
make test-e2e                          # Run backend E2E tests
make test-ui-e2e                       # Run frontend E2E tests

# Cleanup
make nuke                              # Clean up unused Docker images
make add-hooks                         # Enable pre-commit hooks
```

## Testing Best Practices

### E2E Tests

- **Backend**: Test against full development cluster in `test/e2e/`; use helper functions in `test/helpers/`
- **Frontend**: Use Playwright for real browser testing; keep tests in `ui/tests/*.spec.ts`
- **Fixtures**: Define test data fixtures for repeatability and isolation
- **Cleanup**: Always clean up test data after tests complete (delete clusters, tokens, etc.)

### Test Fixtures & Mocks

**Go helpers**: See `test/helpers/*.go` for helper functions used by unit and E2E tests.

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Architecture Deep Dive](ARCHITECTURE.md)
- [Contributing Guide](CONTRIBUTING.md)
- [Security Guidelines](SECURITY.md)
- [Canonical Design System](https://github.com/canonical/react-components)
- [React Query Documentation](https://tanstack.com/query/latest)
- [TanStack Query](https://tanstack.com/query/latest/docs/react/overview)
- [Formik Documentation](https://formik.org/docs/overview)
- [Playwright Testing](https://playwright.dev/)
- [Vite Documentation](https://vitejs.dev/)
- [Vitest Documentation](https://vitest.dev/)
- [React Router Documentation](https://reactrouter.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Goose Migrations](https://github.com/pressly/goose)

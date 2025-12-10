# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Supanova Server is a Go REST API backend for the Supanova Radiation Protection Services learning platform. It uses Echo for HTTP routing, PostgreSQL for data persistence, and sqlc for type-safe database queries.

## Development Commands

### Setup
```bash
make dep              # Download Go module dependencies
```

### Running the Server
```bash
# With Docker (recommended)
docker-compose up -d

# Without Docker (requires Postgres running separately)
docker-compose up -d postgres
make run
```

### Testing
If using Docker with Colima, set these environment variables first:
```bash
export DOCKER_HOST=unix://${HOME}/.colima/default/docker.sock
export TESTCONTAINERS_RYUK_DISABLED=true
```

Then run tests:
```bash
make test             # Run all tests
go test ./internal/tests -run TestIntegration/returns_course_by_id  # Run a single test
```

### Linting
```bash
make lint             # Install golangci-lint and run linter
make lint/fix         # Auto-fix linting issues
```

### Database Operations
```bash
make sqlc                          # Regenerate sqlc code from queries
make migrate/create name=<name>    # Create a new migration file
```

### Building
```bash
make build            # Build Linux binary (CGO_ENABLED=0, GOOS=linux, GOARCH=amd64)
```

## Architecture

### Layer Structure

The codebase follows a clean architecture pattern with clear separation of concerns:

**main.go** → **server** → **handlers** → **domain** ← **store**

1. **main.go**: Application entry point. Handles graceful shutdown, config parsing, database connection, and server lifecycle.

2. **internal/server**: HTTP server setup using Echo framework. Configures middleware (rate limiting: 60 req/min, burst 120), registers routes, and implements custom validator wrapper.

3. **internal/handlers**: HTTP request handlers. Each handler validates input, calls domain repository methods, and returns responses. Dependencies are injected via the `Handlers` struct.

4. **internal/domain**: Domain interfaces and models. Defines repository interfaces (e.g., `CourseRepository`, `ProgressRepository`) that the handlers depend on. This layer contains pure business models, not database models.

5. **internal/store**: Database layer implementing domain repository interfaces. Uses sqlc-generated code for type-safe queries. The `Store` struct embeds sqlc queries and implements all repository interfaces.

6. **internal/config**: Environment variable parsing with validation. Loads from `.env` in development or runtime env vars in production.

### Key Architectural Patterns

- **Dependency Injection**: Handlers receive repository interfaces, not concrete implementations. The main function wires dependencies.

- **Repository Pattern**: Domain defines interfaces, store implements them. This allows the domain layer to be independent of database implementation.

- **sqlc Code Generation**: Database queries are written in `internal/store/queries`, and sqlc generates type-safe Go code in `internal/store/sqlc/`. Never edit generated files directly.

- **Migrations**: Managed by golang-migrate. Migration files live in `internal/store/migrations/`.

### Database Access

- **DO NOT** write raw SQL in handlers or business logic
- **DO** add queries to `internal/store/queries` and run `make sqlc`
- **DO** implement domain repository methods in `internal/store/*.go` using the generated sqlc queries
- **DO NOT** use database models (sqlc.Course) in handlers; use domain models (domain.Course) instead

### Route Registration

Routes are versioned with a `/v2/` prefix and registered in `internal/server/routes.go`. Route-specific registration functions (e.g., `RegisterCourseRoutes`) keep the routing configuration organized.

### Testing

Integration tests use testcontainers to spin up a real Postgres instance. Tests live in `internal/tests/`. The test setup creates a full server instance and makes real HTTP requests.

## Environment Variables

Required environment variables: see `.env.example`


## Important Notes

- The main branch is `main`
- Handlers should NOT directly use sqlc-generated models; convert between sqlc models and domain models in the store layer
- Context is passed through all layers for cancellation and timeout support
- All database queries use pgx/v5 (not database/sql)

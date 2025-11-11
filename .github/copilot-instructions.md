# Copilot Coding Agent Instructions

## Overview

Backend REST API for Mini Maxit - a programming contest platform with automated code evaluation.

**Tech Stack**: Go 1.23+, Gorilla Mux, GORM, PostgreSQL 17, RabbitMQ 3.13
**Size**: 90 Go files in clean architecture (routes → services → repositories)
**Main Entry**: `cmd/app/main.go`

## Structure

```
cmd/app/main.go              # Entry point
internal/
  api/http/routes/           # HTTP handlers (auth, tasks, users, etc.)
  api/http/middleware/       # JWT auth, transaction middleware
  api/queue/                 # RabbitMQ listener
  config/                    # Config from env vars
  database/                  # DB connection setup
  initialization/            # DI container
package/                     # Shared code
  domain/models/             # GORM database models
  domain/schemas/            # API DTOs
  repository/                # Data access (GORM)
  service/                   # Business logic
  utils/                     # Utilities
scripts/
  run-tests.sh              # Test runner (must chmod +x)
  update-docs.sh            # Swagger generator
.golangci.yaml              # Linter config (v2 format!)
```

## Build & Test

### Quick Start
```bash
go mod download              # Download dependencies first
go build -v ./...           # Build all packages (5-10s)
go test -v ./...            # Run tests - NO Docker needed! (1-2s)
golangci-lint run ./...     # Lint (30-60s, must use v2!)
```

### Testing
Tests use mocks - no external dependencies required. Use:
```bash
./scripts/run-tests.sh run      # Run tests
./scripts/run-tests.sh cover    # Generate coverage report
```

Tests are in `*_test.go` files alongside source. Uses `go.uber.org/mock` for mocking.

### Linting - CRITICAL
**MUST use golangci-lint v2**, NOT v1. Config file is v2 format.
```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```
Config is VERY strict (50+ linters). Current status: 0 issues.

### Swagger Docs
After API changes:
```bash
make docs
```

### Docker Dev Environment
Full app needs: PostgreSQL, RabbitMQ, file-storage, worker services.
```bash
docker compose up --build -d
```
Requires `maxit/file-storage` image pre-built. Set `DUMP=true` in `.env` for test users.

## CI/CD Pipeline

### GitHub Actions (all on every push)
1. **go.yaml**: Build (`go build -v ./...`) + Test (`go test -v ./...`) → Docker build/push
2. **pre-commit.yaml**: Runs pre-commit hooks (linting, formatting, tests)
3. **docs.yaml**: Generates Swagger docs on master/develop, publishes to GitHub Pages

### Pre-commit Hooks
- YAML validation, large file check, EOF fixer, trailing whitespace
- `go-fmt`, `go-imports`, `golangci-lint`, `go-unit-tests`
- Swagger doc generation

All must pass for PR merge.

## Configuration

### Required Environment Variables (see `.env.example`)
**Database**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
**API**: `APP_PORT` (default: 8080), `JWT_SECRET_KEY` (required)
**File Storage**: `FILE_STORAGE_HOST`, `FILE_STORAGE_PORT`
**RabbitMQ**: `QUEUE_NAME`, `RESPONSE_QUEUE_NAME`, `QUEUE_HOST`, `QUEUE_PORT`, `QUEUE_USER`, `QUEUE_PASSWORD`
**Other**: `LANGUAGES` (defaults to C/C++), `DUMP` (loads test data), `DEBUG` (loads .env file)

### Key Config Files
- `.golangci.yaml`: v2 format, 50+ strict linters, custom exclusions for docs/tests
- `go.mod`: Go 1.24.2, deps: gorilla/mux, gorm, zap, jwt-go, amqp091-go

## Development Workflow

### Adding API Endpoints
1. Handler in `internal/api/http/routes/<feature>_route.go` with Swagger comments
2. Service logic in `package/service/<feature>_service.go`
3. Repository in `package/repository/<feature>_repository.go` (if DB access needed)
4. Schemas/DTOs in `package/domain/schemas/`
5. Tests in `*_test.go`
6. Run: `go build`, `go test`, `golangci-lint run`, `./scripts/update-docs.sh`

### Change Checklist (ALWAYS in this order)
1. `go build -v ./...` - verify compilation
2. `go test -v ./...` - verify tests pass
3. `golangci-lint run ./...` - check lint (may take 60s)
4. `make generate` - if API changed

### Known Issues
- **Path traversal vulnerability** in `package/utils/file_utils.go` (see TODOs, not being fixed)
- **Script permissions**: Run `chmod +x scripts/*.sh` if "Permission denied"
- **golangci-lint v2 required**: Using v1 will error with "configuration file for v2"

## Architecture

**Clean Architecture**: Routes → Services → Repositories
**Testing**: Use `go.uber.org/mock`, test files alongside source
**DB Migrations**: Auto via GORM AutoMigrate in `internal/database/database.go`
**Logging**: `go.uber.org/zap` - init with `utils.InitializeLogger()`, get logger with `utils.NewNamedLogger("name")`

## Key Files
- `cmd/app/main.go` - Entry point
- `internal/config/config.go` - Env var parsing with defaults
- `internal/database/database.go` - DB connection and migrations
- `internal/initialization/initialization.go` - DI container (~136 lines)
- `internal/api/http/server/server.go` - Route registration

## Validation
These instructions validated by running:
- `go build -v ./...` ✓
- `go test -v ./...` ✓ (all pass)
- `golangci-lint run ./...` ✓ (0 issues)
- Reviewed all workflows and 90 Go files

**Only search further if**: Instructions incomplete for your task, errors not documented here, or specific business logic needed. Otherwise, follow existing patterns for consistency.

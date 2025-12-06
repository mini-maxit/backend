# Generate Test Data Tool

A comprehensive CLI tool for generating test data in the Mini-Maxit database.

## Features

Generates realistic test data for all database entities:
- **Users** with different roles (admin, teacher, student)
- **Groups** with user memberships
- **Tasks** with test cases and file references
- **Contests** with tasks, participants, and registrations
- **Submissions** with results and test outcomes
- **Supporting tables**: AccessControl, ContestParticipant, ContestParticipantGroup, ContestRegistrationRequests, UserGroup, TestCase, File, LanguageConfig, SubmissionResult, TestResult

## Prerequisites

- PostgreSQL database running
- File storage service running (optional, can skip connectivity check)
- Environment variables configured (or use CLI flags)

## Installation

```bash
go build -o generate-test-data ./cmd/generate-test-data
```

## Usage

### Basic Usage

```bash
# Generate default dataset (10 users, 3 groups, 5 tasks, 2 contests)
./generate-test-data

# With environment variables from .env file
export DEBUG=true  # Loads .env file
./generate-test-data
```

### Common Examples

```bash
# Large production-like dataset
./generate-test-data --users 100 --groups 15 --tasks 30 --contests 10 \
  --registration-requests-per-contest 10 --collaborators-per-task 3

# Reproducible data for testing
./generate-test-data --clear-existing --seed 42 --verbose

# Preview what would be created
./generate-test-data --dry-run --users 50 --contests 5

# Skip connectivity checks (for offline testing)
./generate-test-data --skip-connectivity-check
```

## CLI Flags

### User Generation
- `--users int` - Total number of users (default: 10)
- `--admin-count int` - Number of admin users (default: 1)
- `--teacher-count int` - Number of teacher users (default: 2)
- `--student-count int` - Number of student users (0 = remaining, default: 0)
- `--user-password string` - Default password for all users (default: "password123")

### Group Generation
- `--groups int` - Number of groups to create (default: 3)
- `--users-per-group int` - Average number of users per group (default: 5)

### Task Generation
- `--tasks int` - Number of tasks to create (default: 5)
- `--visible-tasks int` - Number of tasks visible to all (default: 3)
- `--tests-per-task int` - Number of test cases per task (default: 3)
- `--fixtures-dir string` - Path to fixtures directory (default: "./fixtures")
- `--create-fixtures` - Create sample fixture files if they don't exist (default: false)

### Contest Generation
- `--contests int` - Number of contests to create (default: 2)
- `--tasks-per-contest int` - Number of tasks per contest (default: 3)
- `--participants-per-contest int` - Individual participants per contest (default: 5)
- `--group-participants-per-contest int` - Group participants per contest (default: 1)

### Submission Generation
- `--submissions-per-task int` - Submissions per standalone task (default: 3)
- `--submissions-per-contest int` - Submissions per contest task (default: 2)

### Supporting Data
- `--registration-requests-per-contest int` - Registration requests per contest (default: 3)
- `--collaborators-per-task int` - Collaborators per task via AccessControl (default: 2)
- `--collaborators-per-contest int` - Collaborators per contest via AccessControl (default: 2)

### Database Configuration
- `--db-host string` - Database host (overrides DB_HOST env var)
- `--db-port int` - Database port (overrides DB_PORT env var)
- `--db-user string` - Database user (overrides DB_USER env var)
- `--db-password string` - Database password (overrides DB_PASSWORD env var)
- `--db-name string` - Database name (overrides DB_NAME env var)

### File Storage Configuration
- `--file-storage-host string` - File storage host (overrides FILE_STORAGE_HOST env var)
- `--file-storage-port string` - File storage port (overrides FILE_STORAGE_PORT env var)

### Utilities
- `--clear-existing` - Truncate all tables before generating new data (default: false)
- `--seed int` - Random seed for reproducible data (0 = time-based, default: 0)
- `--skip-connectivity-check` - Skip pre-flight connectivity checks (default: false)
- `--verbose` - Detailed progress output (default: false)
- `--dry-run` - Show what would be created without actually creating (default: false)

## Pre-flight Checks

Before generating data, the tool verifies:
1. **Database connectivity** - Tests PostgreSQL connection with `sqlDB.Ping()`
2. **File storage availability** - Tests file storage HTTP endpoint

If connectivity fails, you'll see clear error messages:
```
Error: Database connectivity check failed
  Host: localhost:5432
  Error: connection refused
  
Please ensure PostgreSQL is running and accessible.
```

Use `--skip-connectivity-check` to bypass these checks if needed.

## Data Generation Order

The tool generates data in dependency order:
1. LanguageConfig (C, C++, etc.)
2. Users (with bcrypt-hashed passwords)
3. Groups → UserGroup (many-to-many relationships)
4. Tasks → File (descriptions), TestCase → File (test I/O), AccessControl (permissions)
5. Contests → ContestTask, ContestParticipant, ContestParticipantGroup, ContestRegistrationRequests, AccessControl
6. Submissions → File (source), SubmissionResult, TestResult → File (stdout/stderr/diff)

## Examples

### Default Development Dataset
```bash
./generate-test-data
```
Creates:
- 10 users (1 admin, 2 teachers, 7 students)
- 3 groups with ~5 members each
- 5 tasks (3 visible) with 3 test cases each
- 2 contests with 3 tasks and 5 participants each
- 3 submissions per task
- Supporting data (access control, registration requests, etc.)

### Large Production-like Dataset
```bash
./generate-test-data \
  --users 100 \
  --admin-count 5 \
  --teacher-count 15 \
  --groups 10 \
  --tasks 30 \
  --contests 10 \
  --submissions-per-task 10 \
  --registration-requests-per-contest 10
```

### Reproducible CI/Testing Data
```bash
./generate-test-data --clear-existing --seed 42 --verbose
```
Always generates the same data with seed 42.

### Preview Mode
```bash
./generate-test-data --dry-run --users 50 --contests 5
```
Shows configuration without creating any data.

## Troubleshooting

### "Database connectivity check failed"
- Ensure PostgreSQL is running: `docker-compose up -d postgres`
- Check DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME environment variables
- Or use CLI flags: `--db-host localhost --db-port 5432 ...`

### "File storage connectivity check failed"
- Ensure file-storage service is running: `docker-compose up -d file-storage`
- Check FILE_STORAGE_HOST and FILE_STORAGE_PORT environment variables
- Or skip the check: `--skip-connectivity-check`

### "admin + teacher + student counts must equal total users"
- If you specify --student-count, ensure: admin-count + teacher-count + student-count = users
- Or set --student-count 0 to auto-calculate remaining

## Development

The tool is built with:
- **github.com/spf13/cobra** - CLI framework
- **github.com/brianvoe/gofakeit/v7** - Realistic fake data generation
- Existing backend repositories and services - Ensures data consistency with API

All generated data follows the same constraints and validation rules as the API.

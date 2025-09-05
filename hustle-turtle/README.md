# Hustle Turtle

A Go REST API built with Gin framework following Go project layout best practices.

## Features

- Health check endpoint at `/health`
- Database migrations with CLI control
- PostgreSQL support
- Clean architecture with separated concerns
- Environment-based configuration

## Project Structure

```
hustle-turtle/
├── cmd/
│   └── server/           # Application entrypoint
│       └── main.go
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── database/        # Database operations and migrations
│   ├── handlers/        # HTTP handlers
│   └── models/          # Data models
├── pkg/                 # Public library code
│   └── utils/           # Utility functions
├── migrations/          # Database migration files
├── bin/                 # Compiled binaries
├── go.mod
├── go.sum
└── README.md
```

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL database:
```bash
createdb hustle_turtle
```

3. Set environment variables (optional):
```bash
export DATABASE_URL="postgres://username:password@localhost/hustle_turtle?sslmode=disable"
export PORT="8080"
```

**Defaults:**
- `DATABASE_URL`: `postgres://localhost/hustle_turtle?sslmode=disable`
- `PORT`: `8080`

## Running

### Build the application:
```bash
go build -o bin/hustle-turtle ./cmd/server
```

### Start the server (without migrations):
```bash
go run ./cmd/server
# OR using the binary
./bin/hustle-turtle
```

### Start the server with automatic migrations:
```bash
go run ./cmd/server -auto-migrate
# OR using the binary
./bin/hustle-turtle -auto-migrate
```

The server will start on port 8080.

## API Endpoints

- `GET /health` - Health check endpoint

## Database Migrations

Migrations are stored in the `migrations/` directory and can be run separately from the application.

### Migration Commands

#### Run migrations up (apply new migrations):
```bash
go run ./cmd/server -migrate=up
# OR using the binary
./bin/hustle-turtle -migrate=up
```

#### Run migrations down (rollback migrations):
```bash
go run ./cmd/server -migrate=down
# OR using the binary
./bin/hustle-turtle -migrate=down
```

#### Check current migration version:
```bash
go run ./cmd/server -version
# OR using the binary
./bin/hustle-turtle -version
```

### Current Migrations:
- `000001_create_users_table` - Creates users table with email, name, and timestamps

### Migration Best Practices:
- **Development**: Use `-auto-migrate` flag for convenience
- **Production**: Run migrations separately before deploying new code:
  1. `./bin/hustle-turtle -migrate=up` (run migrations)
  2. `./bin/hustle-turtle` (start the service)
- **Rollback**: Use `-migrate=down` if you need to rollback migrations

## Architecture

The project follows the **Standard Go Project Layout**:

- **`cmd/`**: Main applications for this project
- **`internal/`**: Private application and library code
- **`pkg/`**: Library code that's ok to use by external applications
- **`migrations/`**: Database schema migration files

### Internal Package Structure:

- **`config/`**: Configuration management and environment variables
- **`database/`**: Database connection, migrations, and database-related utilities
- **`handlers/`**: HTTP handlers (controllers) for different endpoints
- **`models/`**: Data structures and business logic models

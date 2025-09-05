# Hustle Turtle

A Go REST API built with Gin framework that includes database migrations.

## Features

- Health check endpoint at `/health`
- Automatic database migrations on startup
- PostgreSQL support

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL database:
```bash
createdb hustle_turtle
```

3. Set database URL (optional):
```bash
export DATABASE_URL="postgres://username:password@localhost/hustle_turtle?sslmode=disable"
```

If not set, defaults to: `postgres://localhost/hustle_turtle?sslmode=disable`

## Running

### Start the server (without migrations):
```bash
go run main.go
```

### Start the server with automatic migrations:
```bash
go run main.go -auto-migrate
```

The server will start on port 8080.

## API Endpoints

- `GET /health` - Health check endpoint

## Database Migrations

Migrations are stored in the `migrations/` directory and can be run separately from the application.

### Migration Commands

#### Run migrations up (apply new migrations):
```bash
go run main.go -migrate=up
```

#### Run migrations down (rollback migrations):
```bash
go run main.go -migrate=down
```

#### Check current migration version:
```bash
go run main.go -version
```

### Current Migrations:
- `000001_create_users_table` - Creates users table with email, name, and timestamps

### Migration Best Practices:
- **Development**: Use `-auto-migrate` flag for convenience
- **Production**: Run migrations separately before deploying new code:
  1. `go run main.go -migrate=up` (run migrations)
  2. `go run main.go` (start the service)
- **Rollback**: Use `-migrate=down` if you need to rollback migrations

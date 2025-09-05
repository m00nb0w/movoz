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

```bash
go run main.go
```

The server will start on port 8080 and run migrations automatically.

## API Endpoints

- `GET /health` - Health check endpoint

## Migrations

Migrations are stored in the `migrations/` directory and run automatically on startup.

Current migrations:
- `000001_create_users_table` - Creates users table with email, name, and timestamps

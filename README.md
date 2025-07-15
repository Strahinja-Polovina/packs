# Packs API Service

A Go microservice for optimal pack combination calculations and order management.

## Architecture

Built using Clean Architecture principles:

```
┌─────────────────────────────────────┐
│         Presentation Layer          │
│      (HTTP Handlers, Routes)        │
├─────────────────────────────────────┤
│         Application Layer           │
│              (Services)             │
├─────────────────────────────────────┤
│          Domain Layer               │
│      (Entities, Repositories)       │
├─────────────────────────────────────┤
│        Infrastructure Layer         │
│       (Database, External APIs)     │
└─────────────────────────────────────┘
```

**Project Structure:**
```
├── cmd/api/              # Application entry point
├── internal/
│   ├── domain/           # Core business logic
│   ├── application/      # Use cases and services
│   ├── infrastructure/   # External concerns
│   └── presentation/     # HTTP handlers & templates
│       ├── handlers/     # HTTP request handlers
│       ├── routes/       # Route definitions
│       └── templates/    # Templ templates for frontend
├── pkg/                 # Shared utilities
├── migrations/          # Database migrations
└── docs/               # API documentation
```

## Versions

- **Go**: 1.24+
- **PostgreSQL**: 17+
- **Docker**: 27+

## Data Flow

```
HTTP Request → Handler → Service → Repository → Database
Database → Repository → Service → Handler → HTTP Response
```

## Quick Start

### With Docker (Recommended)
```bash
# Start all services
make up

# Access the application
open http://localhost:8080/                    # Frontend interface
open http://localhost:8080/swagger/index.html  # API documentation
```

### Without Docker
```bash
# Setup environment
make setup

# Create database
createdb packs_db

# Run migrations
make migrate-up

# Start application
make run
```

## Commands

```bash
# Development
make help                # Show all commands
make build              # Build application
make run                # Run locally
make test               # Run tests
make test-coverage      # Run tests with coverage
make lint               # Check code quality
make templ-generate     # Generate templ templates

# Database
make migrate-up         # Apply migrations
make migrate-down       # Rollback migration
make migrate-status     # Check migration status

# Docker
make docker-build       # Build Docker image
make docker-run         # Run Docker container
make docker-compose-up  # Start with docker-compose
make docker-compose-down # Stop services

# Documentation
make swagger            # Generate API docs
```

## Access Points

Once running, you can access:

### Frontend Interface
**http://localhost:8080/**
- Web-based interface for pack management
- Order creation and management
- Interactive package configuration

### API Documentation
**http://localhost:8080/swagger/index.html**
- Interactive API documentation
- Test API endpoints directly
- View request/response schemas

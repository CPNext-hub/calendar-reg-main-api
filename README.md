# calendar-reg-main-api

Calendar Registration Main API — built with Go + Fiber, following Clean Architecture.

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration (env vars)
│   ├── domain/
│   │   ├── entity/              # Business entities
│   │   │   ├── app_info.go
│   │   │   └── health.go
│   │   └── usecase/             # Business logic / use cases
│   │       ├── health_usecase.go
│   │       └── version_usecase.go
│   └── delivery/
│       └── http/
│           ├── handler/         # HTTP handlers (controllers)
│           │   ├── health_handler.go
│           │   └── version_handler.go
│           ├── middleware/      # HTTP middlewares
│           │   └── middleware.go
│           ├── router/          # Route definitions
│           │   └── router.go
│           └── server/          # Server bootstrap
│               └── server.go
├── .env.example
├── Makefile
├── go.mod
└── go.sum
```

## Getting Started

```bash
# Install dependencies
go mod tidy

# Run the server
make run
# or
go run cmd/api/main.go
```

## API Endpoints

| Method | Path               | Description             |
|--------|--------------------|-------------------------|
| GET    | `/api/v1/status`   | Health check            |
| GET    | `/api/v1/version`  | App name, version, env  |

## Configuration

Set via environment variables (or `.env` file):

| Variable      | Default                  |
|---------------|--------------------------|
| `APP_NAME`    | calendar-reg-main-api    |
| `APP_VERSION` | 0.1.0                    |
| `APP_ENV`     | development              |
| `PORT`        | 8080                     |
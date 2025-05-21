# Vertical Slice Architecture

## Application Usage

### Install Dependencies

These docs assume the brew is already installed on your system. For more information on installing go-task, see the instructions [here](https://taskfile.dev/installation/).

```bash
# Install Go Task
brew install go-task/tap/go-task
```

All other dependencies can be installed with:

```bash
task setup:deps
```

### Run the Application

- Run database and migrations

    ```bash
    task db:start
    ```

## Architecture Explanation

### File Structure

```text
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── users/
│   │   ├── get_handler.go
│   │   ├── create_handler.go
│   │   ├── update_handler.go
│   │   ├── delete_handler.go
│   │   ├── service.go
│   │   ├── routes.go
│   │   └── models.go
│   ├── comments/
│   │   ├── get_handler.go
│   │   ├── create_handler.go
│   │   ├── update_handler.go
│   │   ├── delete_handler.go
│   │   ├── service.go
│   │   ├── routes.go
│   │   └── models.go
│   └── middleware/
│       ├── middleware.go
│       ├── recovery.go
│       └── logger.go
├── go.mod
└── go.sum
```

### General Notes

- Vertical slice architecture is a natural evolution of layered architecture for situations where
  there is more than one resource/entity type in the application. It is most appropriate for
  monolithic applications rather than microservices.
- With vertical slice architecture, each resource/entity type has its own package that contains
  all the logic for that resource/entity type. This includes handlers, services, and models. Any
  shared logic can be placed in separate packages.

### Example Applications

- Monolithic applications
- Applications with multiple resource/entity types

### Comparison

#### Pros

- Provides a clear separation of concerns for each resource/entity type in the application.
- Can scale well as the application grows in complexity in a way that simpler architectures cannot.

#### Cons

- The use of resource/entity type specific packages when there is a single resource/entity type can
  be too complex and App architecture may be more appropriate.
# App Package Architecture

## Architecture Explanation

### File Structure

```text
.
├── cmd/
│   └── app/
│       └── main.go                # Application entry point
├── internal/
│   └── app/
│       ├── app.go                 # Handler setup and response encoding utilities
│       ├── config.go              # Application configuration loading from environment
│       ├── routes.go              # Route registration and HTTP handler wiring
│       ├── models.go              # User model and related types
│       ├── middleware.go          # Middleware for logging, tracing, etc.
│       ├── create_user.go         # Handler: Create a new user (POST /user)
│       ├── read_user.go           # Handler: Get a user by ID (GET /user/{id})
│       ├── update_user.go         # Handler: Update a user by ID (PUT /user/{id})
│       ├── delete_user.go         # Handler: Delete a user by ID (DELETE /user/{id})
│       ├── list_users.go          # Handler: List all users (GET /user)
├── db/
│   ├── migrations/                # Database schema migrations and seed data
│   └── conf/                      # Database migration tool configuration
├── tests/
│   └── integration/
│       ├── integration_test.go    # Integration tests for user CRUD endpoints
│       └── helper.go              # Test helpers (in-memory DB, test server setup)
├── go.mod
├── go.sum
├── Taskfile.yml                   # Task runner configuration for automation
├── .golangci.yaml                 # GolangCI-Lint configuration
├── docker-compose.yaml            # Docker Compose configuration for local dev/test
└── dockerfile                     # Dockerfile for building the application container
```

### General Notes

- The App Package architecture is a natural progression from a completely flat architecture. It
  introduces a minimal but meaningful structure, making it easy to understand and work with.
- This architecture is intentionally simple, focusing on clarity and ease of use rather than
  enforcing strict boundaries between business logic and data access. This can speed up development
  for small to medium-sized projects.
- The use of a `cmd` directory signals that the project is meant to be run as an application (not
  just a library), and the `internal` directory ensures that implementation details are hidden from
  external packages, promoting encapsulation and preventing accidental imports. This architecture
  does allow for more entrypoints (e.g. CLI) by creating more entrypoints in the `cmd` directory.
- While the architecture does not enforce a domain-driven or layered approach, it provides enough
  separation to keep code organized and understandable, especially for teams or new contributors.
- This structure is ideal for projects where rapid iteration and simplicity are more important than
  strict adherence to architectural patterns. It is often used as a starting point before evolving
  into more complex architectures as requirements grow.
- As a general rule, the App Package architecture is the simplest architecture that should be used
  for production code, balancing maintainability and minimalism.
- This example is missing some components present in other architecture examples in this repo (caching, otel, etc). This is intentional. Since the app-architecture design is designed to be fairly simple, it wouldn't make sense to have all the bells and whistles in an example application.

### Example Applications

- **AWS Lambda Functions:**  
  Suitable for serverless applications where a single handler manages REST API requests or processes
  events from sources like SQS, DynamoDB Streams, or SNS. The simple structure keeps deployment and
  maintenance straightforward.
- **Simple CLI Applications:**  
  Ideal for command-line tools that require a clear entry point and a small set of internal logic,
  without the overhead of a more complex architecture.
- **Microservices:**  
  Works well for small to medium-sized microservices that expose a REST API or gRPC endpoints,
  especially when each service is focused on a single domain or responsibility.
- **Internal Tools and Utilities:**  
  Great for building internal dashboards, automation scripts, or batch processing jobs where rapid
  development and ease of understanding are priorities.
- **Prototyping and MVPs:**  
  Perfect for proof-of-concept projects or MVPs where requirements may change rapidly and the
  overhead of a more layered architecture is not justified.

### Comparison

#### Pros

- **Simplicity:**  
  The project structure is easy to understand, making onboarding new developers faster and reducing
  cognitive overhead.
- **Clarity:**  
  Directory and file naming conventions (`cmd`, `internal`, etc.) provide clear guidance on the
  intended use and boundaries of code, reducing ambiguity.
- **Focus:**  
  Developers can concentrate on implementing application features without being distracted by
  architectural complexity or boilerplate.
- **Encapsulation:**  
  The `internal` directory prevents accidental use of internal code by other projects, enforcing a
  clean separation between public and private APIs.
- **Reduced Risk of Circular Dependencies:**  
  The flat but organized structure minimizes the chances of introducing circular dependencies, which
  can be a common issue in larger Go projects.
- **Easy Refactoring:**  
  The minimal structure makes it easier to refactor or evolve the architecture as the application
  grows.

#### Cons

- **Limited Scalability for Complex Projects:**  
  As the application grows in size and complexity, the lack of additional directories or layers (
  such as `service`, `repository`, or `domain`) can make it harder to manage and maintain.
- **Potential for Mixing Concerns:**  
  Without strict boundaries, business logic and data access code may become intertwined, making
  testing and future refactoring more difficult.
- **Not Ideal for Large Teams:**  
  In larger teams or organizations, the lack of enforced separation can lead to inconsistent code
  organization and technical debt.
- **Difficult to Enforce Best Practices:**  
  The architecture relies on developer discipline rather than structural enforcement, which can lead
  to deviations from best practices over time.
- **Migration Overhead:**  
  If the project outgrows this architecture, migrating to a more layered or modular structure may
  require significant refactoring.

## Application Usage

### Install Dependencies

These docs assume the brew is already installed on your system. For more information on installing
go-task, see the instructions [here](https://taskfile.dev/installation/).

```bash
# Install Go Task
brew install go-task/tap/go-task
```

All other dependencies can be installed with:

```bash
task setup:deps
```

### Run the Application

- Build Docker Image

  ```bash
  task docker:build
  ```

- Run the Application Locally

  ```bash
  task app:start
  ```

- Stop Docker Images

  ```bash
  task docker:stop
  ```

- Run database and migrations

    ```bash
    task db:start
    ```

### Working Locally

- Run Unit  Tests

  ```bash
  task test:unit
  ```

- Run Integration Tests

  ```bash
  task test:integration
  ```

- Format Go Files

  ```bash
  task format
  ```

You can see other task commands that are available by running `task`
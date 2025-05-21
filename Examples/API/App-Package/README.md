# App Package Architecture

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
│   └── app/
│       └── main.go
├── internal/
│   └── app/
│       ├── routes.go
│       ├── get_user.go
│       ├── create_user.go
│       ├── update_user.go
│       ├── delete_user.go
│       ├── models.go
│       └── middleware.go
├── go.mod
└── go.sum
```

### General Notes

- The App application architecture is a natural progression from a completely flat architecture. It is a
  simple
  way to organize a project that is more scalable than the flat architecture.
- It does not enforce a strict separation of concerns between business logic and data access logic.
- The use of a `cmd` directory indicates that the project is for an application that is intended to
  be
  run as a standalone application. Similarly, the use of an `internal` directory indicates that any
  code inside of it is not intended to be imported by other packages.
- As a general rule, the App architecture is the simplest architecture that should be used for
  production code.

### Example Applications

- AWS Lambda with a single handler for a REST API or an event driven application using SQS/DynamoDB
  Stream/etc.
- Simple CLI Application

### Comparison

#### Pros

- The project is simple and easy to understand.
- Provides more information to others based on the naming of folders than a flat project structure.
- The use of `cmd` and `internal` directories provides a clear indication of the intended use of the
  code in the project.
- Allows you to focus on the functionality of your application without worrying about the
  complexities of other architectures.
- Eliminates the risk of circular dependencies

#### Cons

- The lack of any other directories besides `cmd/app` and `internal/app` can make it difficult to
  use on more complex applications.
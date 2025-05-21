# Layered Architecture

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
│   ├── routes/
│   │   └── routes.go
│   ├── handlers/
│   │   ├── request.go
│   │   ├── response.go
│   │   ├── get_user.go
│   │   ├── create_user.go
│   │   ├── update_user.go
│   │   └── delete_user.go
│   ├── services/
│   │   └── users.go
│   ├── models/
│   │   └── models.go
│   └── middleware/
│       ├── middleware.go
│       ├── recovery.go
│       └── logger.go
├── go.mod
└── go.sum
```

### General Notes

- Layered architecture is one of the most common Go application architectures and is arguably
  closest to the architectures used in most enterprise applications.
- It is most appropriate to be used for microservices that need to handle multiple actions on a
  single resource. As an example, a REST API that performs CRUD actions on a single entity type.
- The lack of domain specific names for packages is due to the project being the domain context as a
  whole.

### Example Applications

- REST API microservice
- Event driven microservice
- Medium to high complexity CLI

### Comparison

#### Pros

- Organization provides great scalability for the application and its clear where each piece of the
  application goes.
- Functionality in each package can be exported or un-exported depending on the situation providing
  separation of concerns, especially when used with inversion of control.
- Can be more comfortable for developers from other languages who are new to Go.

#### Cons

- The package naming convention can be too rigid, and it can encourage more formulaic development
  that does not properly take into account all aspects of a given application.
- It can be over complex for simpler applications.
# App Package

## Install Dependencies

Assume brew installed

```bash
# Install Go
brew install go

# install colima
brew install colima

# Install Go Task
brew install go-task/tap/go-task

# Install mockery
go install github.com/vektra/mockery/v3@v3.2.5

# install golangci-lint
brew install golangci-lint
```

## Run the Application

- Run database and migrations

    ```bash
    task docker-db-up
    ```
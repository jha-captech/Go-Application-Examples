package integration

import (
	"fmt"
	"log/slog"
	"net/http/httptest"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3" // import SQLite driver for sqlx

	"example.com/examples/api/app-package/internal/app"
)

// newTestDB creates and returns an in-memory SQLite database
// pre-populated with a 'users' table and some test user data.
// This is useful for integration tests that require a database.
func newTestDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory db: %w", err)
	}

	schema := `
    CREATE TABLE users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );
    INSERT INTO users (name, email, password) VALUES
        ('Alice', 'alice@example.com', 'password123'),
        ('Bob', 'bob@example.com', 'securepass456'),
        ('Carol', 'carol@example.com', 'carolpass789'),
        ('Dave', 'dave@example.com', 'davepass321');
    `

	// execute schema creation and data insertion
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema or insert test data: %w", err)
	}

	return db, nil
}

// newTestServer creates a new httptest.Server instance using the app's HTTP handler,
// with middleware applied, and an in-memory test database. It returns the server,
// the database connection, and any error encountered. Useful for end-to-end HTTP integration tests.
func newTestServer() (*httptest.Server, *sqlx.DB, error) {
	// set up in-memory database
	db, err := newTestDB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test database: %w", err)
	}

	logger := slog.Default()

	// create handler and wrap in middleware
	handler := app.NewHandler(logger, db)
	wrappedHandler := app.WrapHandler(
		handler,
		app.TraceIDMiddleware(),
		app.LoggingMiddleware(logger),
		app.RecoveryMiddleware(logger),
	)

	// start a test server with wrapped handler
	server := httptest.NewServer(wrappedHandler)

	return server, db, nil
}

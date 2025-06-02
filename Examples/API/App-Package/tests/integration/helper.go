package integration

import (
	"fmt"
	"log/slog"
	"net/http/httptest"

	"example.com/examples/api/app-package/internal/app"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// NewTestDB returns an in-memory SQLite DB with a users table and some test data.
func newTestDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory db: %v", err)
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

	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema or insert test data: %v", err)
	}

	return db, nil
}

func newTestServer() (*httptest.Server, *sqlx.DB, error) {
	// set up in-memory database
	db, err := newTestDB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test database: %v", err)
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

	// Start a test server with your wrapped handler
	server := httptest.NewServer(wrappedHandler)
	return server, db, nil
}

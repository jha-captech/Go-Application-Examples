package integration

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	_ "github.com/mattn/go-sqlite3"

	"example.com/examples/api/layered/internal/middleware"
	"example.com/examples/api/layered/internal/routes"
	"example.com/examples/api/layered/internal/services"
	"example.com/examples/api/layered/internal/telemetry"
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

type TestRedis struct{}

func (r *TestRedis) Set(
	ctx context.Context,
	key string,
	value any,
	exp time.Duration,
) *redis.StatusCmd {
	return redis.NewStatusCmd(ctx, "OK")
}

func (r *TestRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	return redis.NewStringCmd(ctx, redis.Nil)
}

func (r *TestRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return redis.NewIntCmd(ctx, int64(len(keys)))
}

func (r *TestRedis) Ping(ctx context.Context) *redis.StatusCmd {
	return redis.NewStatusCmd(ctx, "OK")
}

func newTestServer() (*httptest.Server, *sqlx.DB, error) {
	// set up in-memory database
	db, err := newTestDB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test database: %v", err)
	}

	logger := slog.Default()

	rdb := &TestRedis{}

	// Create a new users service
	usersService := services.NewUsersService(logger, db, rdb, 0)

	// Create a serve mux to act as our route multiplexer
	mux := telemetry.InstrumentServeMux(http.NewServeMux())

	// Add our routes to the mux
	routes.AddRoutes(mux, logger, usersService)

	// create handler and wrap in middleware
	wrappedMux := middleware.WrapHandler(
		mux.InstrumentRootHandler(),
		middleware.TraceID(),
		middleware.Logger(logger),
		middleware.Recover(logger),
	)

	// Start a test server with your wrapped handler
	server := httptest.NewServer(wrappedMux)
	return server, db, nil
}

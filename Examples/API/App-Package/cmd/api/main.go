package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"example.com/examples/api/app-package/internal/app"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Load and validate environment config
	cfg, err := app.NewConfig()
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load config: %w", err)
	}

	// Create a structured logger, which will print logs in json format to the
	// writer we specify.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	// Create a new DB connection using environment config
	logger.DebugContext(ctx, "Connecting to database")
	db, err := sql.Open("pgx", fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBUserName,
		cfg.DBUserPassword,
		cfg.DBName,
		cfg.DBPort,
	))
	if err != nil {
		return fmt.Errorf("[in main.run] failed to open database: %w", err)
	}

	// Ping the database to verify connection
	logger.DebugContext(ctx, "Pinging database")
	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("[in main.run] failed to ping database: %w", err)
	}

	defer func() {
		logger.DebugContext(ctx, "Closing database connection")
		if err = db.Close(); err != nil {
			logger.ErrorContext(ctx, "Failed to close database connection", "err", err)
		}
	}()

	logger.InfoContext(ctx, "Connected successfully to the database")

	handler := app.NewHandler(logger, db)

	// Create a new http server with our mux as the handler
	// Create a new http server with our mux as the handler
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	fmt.Println("Starting server on", httpServer.Addr)

	if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("[in main.run] failed to listen and serve: %w", err)
	}

	return nil
}

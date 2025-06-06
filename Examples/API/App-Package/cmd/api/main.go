package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"

	_ "github.com/jackc/pgx/v5/stdlib"

	"example.com/examples/api/app-package/internal/app"
)

func main() {
	// Run the main application logic and handle any errors.
	if err := run(context.Background()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Load and validate environment config from env vars or files.
	cfg, err := app.NewConfig()
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load config: %w", err)
	}

	// Create a structured logger that outputs JSON logs to stdout.
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout, &slog.HandlerOptions{
				Level: cfg.LogLevel,
			},
		),
	)

	// Connect to the PostgreSQL database using the provided config.
	logger.DebugContext(ctx, "Connecting to and pinging the database")

	db, err := sqlx.Connect(
		"pgx", fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			cfg.DBHost,
			cfg.DBUserName,
			cfg.DBUserPassword,
			cfg.DBName,
			cfg.DBPort,
		),
	)
	// Ensure the database connection is closed on exit.
	defer func() {
		logger.DebugContext(ctx, "Closing database connection")

		if err = db.Close(); err != nil {
			logger.ErrorContext(ctx, "Failed to close database connection", "err", err)
		}
	}()

	// Create the main HTTP handler and wrap it with middleware for tracing, logging, and recovery.
	handler := app.NewHandler(logger, db, cfg.EnableSwagger)

	// Wrap the handler with middleware for tracing, logging, and recovery.
	wrappedHandler := app.WrapHandler(
		handler,
		app.TraceIDMiddleware(),
		app.LoggingMiddleware(logger),
		app.RecoveryMiddleware(logger),
	)

	// Create the HTTP server with the wrapped handler.
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: wrappedHandler,
	}

	// Set up context that cancels on SIGINT or SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Use errgroup to manage goroutines and propagate errors.
	eg, ctx := errgroup.WithContext(ctx)

	// Register a shutdown function to be called when the context is cancelled.
	context.AfterFunc(
		ctx,
		func() {
			eg.Go(
				func() error {
					// Attempt graceful shutdown of the HTTP server.
					if err := httpServer.Shutdown(ctx); err != nil {
						return fmt.Errorf("[in main.run] failed to shutdown server: %w", err)
					}

					return nil
				},
			)
		},
	)

	// Start the HTTP server.
	logger.InfoContext(ctx, "Starting HTTP server", "addr", httpServer.Addr)

	// Listen and serve; return error unless it's a normal server close.
	if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("[in main.run] failed to listen and serve: %w", err)
	}

	return nil
}
